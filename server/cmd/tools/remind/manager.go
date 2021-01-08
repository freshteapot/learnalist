package remind

import (
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/mobile"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
)

var managerCMD = &cobra.Command{
	Use:   "manager",
	Short: "Reminder manager",
	Long: `
kubectl port-forward svc/nats 4222:4222 &

TOPIC=lal.monolog \
EVENTS_STAN_CLIENT_ID=remind-daily-in \
EVENTS_STAN_CLUSTER_ID=test-cluster \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools remind manager
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		logContext := logger.WithField("context", "remind")

		event.SetDefaultSettingsForCMD()
		event.SetupEventBus(logContext)

		viper.SetDefault("topic", event.TopicMonolog)
		viper.BindEnv("topic", "TOPIC")

		subjectRead := viper.GetString("topic")

		databaseName := viper.GetString("remind.daily.sqlite.database")
		db := database.NewDB(databaseName)

		remindDailyRepo := remind.NewRemindDailySettingsSqliteRepository(db)
		mobileRepo := mobile.NewSqliteRepository(db)
		dailyManager := remind.NewDaily(
			remindDailyRepo,
			mobileRepo,
			logger.WithField("context", "daily-reminder"))

		userStorageRepo := userStorage.NewSqliteManagementStorage(db)
		spacedRepetitionRepo := spaced_repetition.NewSqliteRepository(db)
		remindSpacedRepetitionRepo := remind.NewRemindSpacedRepetitionSqliteRepository(db)
		spacedRepetitionManager := remind.NewSpacedRepetition(
			userStorageRepo,
			spacedRepetitionRepo,
			remindSpacedRepetitionRepo,
			logger.WithField("context", "spaced-repetition-reminder"))

		sc := event.GetBus().(*event.NatsBus).Connection()

		subscribers := make([]event.NatsSubscriber, 0)
		subscribers = append(subscribers, dailyManager)
		subscribers = append(subscribers, spacedRepetitionManager)

		for _, subscriber := range subscribers {
			err := subscriber.Subscribe(subjectRead, sc)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to start subscriber in remind manager")
			}
		}

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
			break
		}

		for _, subscriber := range subscribers {
			subscriber.Close()
		}

		event.GetBus().Close()
	},
}
