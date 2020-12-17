package remind

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
)

var dailyCMD = &cobra.Command{
	Use:   "daily",
	Short: "Consume for daily reminders",
	Long: `

	ssh $SSH_SERVER -L 4222:127.0.0.1:4222 -N &
	ssh $SSH_SERVER sudo kubectl port-forward deployment/stan01 4222:4222 &

	TOPIC=lal.monolog \
	EVENTS_STAN_CLIENT_ID=remind-daily-in \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run main.go --config=../config/dev.config.yaml \
	tools remind daily


	I can reuse mobile-device for daily_reminder_medium_push
	`,
	Run: func(cmd *cobra.Command, args []string) {
		startAt := time.Now().UTC()
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		viper.SetDefault("topic", "lal.monolog")
		viper.BindEnv("topic", "TOPIC")
		subjectRead := viper.GetString("topic")

		natsServer := viper.GetString("server.events.nats.server")
		clusterID := viper.GetString("server.events.stan.clusterID")
		clientID := viper.GetString("server.events.stan.clientID")
		opts := []nats.Option{nats.Name("lal-go-server")}
		nc, err := nats.Connect(natsServer, opts...)

		if err != nil {
			panic(err)
		}

		logContext := logger.WithFields(logrus.Fields{
			"context":    "daily-in",
			"cluster_id": clusterID,
			"client_id":  clientID,
		})

		databaseName := viper.GetString("remind.daily.sqlite.database")
		db := database.NewDB(databaseName)

		settingsRepo := remind.NewRemindDailySettingsSqliteRepository(db)
		mobileRepo := mobile.NewSqliteRepository(db)

		manager := NewManager(settingsRepo, mobileRepo, logContext)
		fmt.Println("mobileRepo", mobileRepo)
		fmt.Println("settingsRepo", settingsRepo)

		logContext.Info("Connecting to nats server...")
		sc, err := stan.Connect(clusterID, clientID,
			stan.NatsConn(nc),
			stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
				logContext.Fatalf("Connection lost, reason: %v", reason)
			}),
			stan.Pings(10, 5),
		)

		if err != nil {
			logContext.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, nc.Opts.Url)
		}
		defer logCloser(sc)

		allowed := []string{
			mobile.EventMobileDeviceRemoved,
			mobile.EventMobileDeviceRegistered,
			remind.EventApiRemindDailySettings,
			event.ApiSpacedRepetition,
			plank.EventApiPlank,
		}

		d := 200 * time.Millisecond
		// Initially we shall wait

		var timer *time.Timer
		timer = time.AfterFunc(500*time.Millisecond, func() {
			manager.SendNotifications(startAt)
			timer.Stop()
		})
		defer timer.Stop()

		handle := func(msg *stan.Msg) {
			var moment event.Eventlog
			json.Unmarshal(msg.Data, &moment)

			if !utils.StringArrayContains(allowed, moment.Kind) {

				timer.Reset(d)
				return
			}

			manager.Write(moment)
			timer.Reset(d)
		}

		durableName := "remind.daily"
		subscription, _ := sc.Subscribe(
			subjectRead,
			handle,
			stan.DurableName(durableName),
			stan.DeliverAllAvailable(),
			stan.MaxInflight(1),
		)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
			break
		}

		err = subscription.Close()
		if err != nil {
			logContext.WithField("error", err).Error("closing subscription")
		}

		err = sc.Close()
		if err != nil {
			logContext.WithField("error", err).Error("closing stan")
		}
	},
}

func logCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close error: %s", err)
	}
}

type manager struct {
	settingsRepo remind.RemindDailySettingsRepository
	mobileRepo   mobile.MobileRepository
	logContext   logrus.FieldLogger
}

func NewManager(settingsRepo remind.RemindDailySettingsRepository, mobileRepo mobile.MobileRepository, logContext logrus.FieldLogger) *manager {
	return &manager{
		settingsRepo: settingsRepo,
		mobileRepo:   mobileRepo,
		logContext:   logContext,
	}
}

// TODO now we need token
// TODO now we need display_name
// daily_reminder_medium_push
// daily_reminder_settings
func (m *manager) Write(entry event.Eventlog) {
	switch entry.Kind {
	case mobile.EventMobileDeviceRemoved:
		fmt.Println("remove token from daily_reminder_medium_push")
	case mobile.EventMobileDeviceRegistered:
		m.processMobileDeviceRegistered(entry)
	case event.ApiSpacedRepetition:
		fmt.Println("Set event added happened=1")
	case plank.EventApiPlank:
		fmt.Println("Set event added happened=1")
	case remind.EventApiRemindDailySettings:
		m.processSettings(entry)
	default:
		return
	}
}

func (m *manager) SendNotifications(startAt time.Time) {
	// Process queue from oldeset to newest
	// Lookup
	// SELECT * FROM daily_reminders
	// Send
	// Update + set event happened=0
	fmt.Println(startAt.Format(time.RFC3339Nano))
}

func (m *manager) whenNext(from time.Time, to time.Time) time.Time {
	whenNext := to
	if to.Before(from) {
		whenNext = to.Add(time.Duration(24 * time.Hour))
	}
	return whenNext
}

func (m *manager) processSettings(entry event.Eventlog) {
	b, _ := json.Marshal(entry.Data)
	var moment event.EventKV
	json.Unmarshal(b, &moment)
	b, _ = json.Marshal(moment.Data)
	var pref remind.UserPreference
	json.Unmarshal(b, &pref.DailyReminder)

	userUUID := moment.UUID
	conf := pref.DailyReminder.RemindV1
	// action = delete = remove
	if entry.Action == event.ActionDeleted {
		fmt.Println("Remove settings from db")

		err := m.settingsRepo.DeleteByUserAndApp(userUUID, conf.AppIdentifier)
		if err != nil {
			m.logContext.Error("failed to remove settings")
		}
		return
	}
	// action = upsert = save
	if entry.Action == event.ActionUpsert {
		conf := pref.DailyReminder.RemindV1

		loc, _ := time.LoadLocation(conf.Tz)
		// Settings + when_next (utc)
		// Take settings
		// Calculate next entry, add it
		parts := strings.Split(conf.TimeOfDay, ":")
		hour, _ := strconv.Atoi(parts[0])
		minute, _ := strconv.Atoi(parts[1])
		now := time.Now()
		local := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
		// Has the time already passed, if yes add date
		whenNext := m.whenNext(now, local)

		fmt.Printf("user:%s tz:%s timeOfDay:%s\n", userUUID, conf.Tz, conf.TimeOfDay)
		fmt.Println("is before: ", local.Before(now))
		fmt.Println("utc ", now.UTC().Format(time.RFC3339Nano))
		fmt.Println("whenNext local ", whenNext.Format(time.RFC3339Nano))
		fmt.Println("whenNext utc ", whenNext.UTC().Format(time.RFC3339Nano))
		fmt.Printf("Write config to db and set when next to %s (utc: %s)\n", whenNext.Format(time.RFC3339Nano), whenNext.UTC().Format(time.RFC3339Nano))

		// I wonder if this can be cast
		err := m.settingsRepo.Save(userUUID,
			openapi.RemindDailySettings{
				TimeOfDay:     conf.TimeOfDay,
				Tz:            conf.Tz,
				Medium:        conf.Medium,
				AppIdentifier: conf.AppIdentifier,
			},
			whenNext.UTC().Format(time.RFC3339Nano),
		)
		if err != nil {
			m.logContext.Error("failed to save settings")
		}
		return
	}
	return
}

func (m *manager) processMobileDeviceRegistered(entry event.Eventlog) {
	var momentKV event.EventKV
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &momentKV)
	b, _ = json.Marshal(momentKV.Data)
	var moment mobile.DeviceInfo
	json.Unmarshal(b, &moment)

	_, err := m.mobileRepo.SaveDeviceInfo(moment.UserUUID, openapi.HttpMobileRegisterInput{
		Token:         moment.Token,
		AppIdentifier: moment.AppIdentifier,
	})

	if err != nil {
		m.logContext.Error("failed to save mobile device")
	}
}
