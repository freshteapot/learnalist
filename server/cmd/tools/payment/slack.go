package payment

import (
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/payment"
	"github.com/slack-go/slack"
)

var slackCMD = &cobra.Command{
	Use:   "slack",
	Short: "Read events and write to slack",
	Run: func(cmd *cobra.Command, args []string) {
		// Listen for all events to payments topic
		// eventID + type

		logger := logging.GetLogger()
		logContext := logger.WithField("context", "payment-slack")

		os.Setenv("EVENTS_STAN_CLIENT_ID", "tools-payment-slack")
		event.SetDefaultSettingsForCMD()
		event.SetupEventBus(logContext)

		viper.SetDefault("topic", event.TopicPayments)
		viper.BindEnv("topic", "TOPIC")

		viper.SetDefault("payment.slack.webhook", "")
		viper.BindEnv("payment.slack.webhook", "PAYMENT_SLACK_WEBHOOK")
		webhook := viper.GetString("payment.slack.webhook")
		if webhook == "" {
			panic("Webhook shouldnt be empty")
		}

		databaseName := viper.GetString("payment.sqlite.database")
		db := database.NewDB(databaseName)

		repo := payment.NewSqliteRepository(db)
		slackListener := payment.NewSlackListener(slack.PostWebhook, webhook, logContext)
		paymentManager := payment.NewManagerListener(repo, logContext)

		subjectRead := viper.GetString("topic")
		sc := event.GetBus().(*event.NatsBus).Connection()

		subscribers := make([]event.NatsSubscriber, 0)
		subscribers = append(subscribers, slackListener)
		subscribers = append(subscribers, paymentManager)

		for _, subscriber := range subscribers {
			err := subscriber.Subscribe(subjectRead, sc)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to start subscriber")
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
