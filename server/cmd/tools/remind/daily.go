package remind

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
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
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		event.SetupEventBus(logger.WithField("context", "remind-daily"))

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

		manager := remind.NewManager(db, settingsRepo, mobileRepo, logContext)
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

		allowed := manager.FilterKindsBy()

		d := 200 * time.Millisecond
		// Initially we shall wait

		var timer *time.Timer
		timer = time.AfterFunc(500*time.Millisecond, func() {
			manager.StartSendNotifications()
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
