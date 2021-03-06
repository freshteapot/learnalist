package tools

import (
	"os"
	"os/signal"

	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	eventReader "github.com/freshteapot/learnalist-api/server/pkg/event/slack"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
)

var slackEventsCMD = &cobra.Command{
	Use:   "slack-events",
	Short: "Read events and write to slack",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		logger.Info("Read events")
		event.SetDefaultSettingsForCMD()
		event.SetupEventBus(logger.WithField("context", "event-bus-setup"))

		viper.SetDefault("server.events.slack.webhook", "")
		viper.BindEnv("server.events.slack.webhook", "EVENTS_SLACK_WEBHOOK")
		webhook := viper.GetString("server.events.slack.webhook")
		if webhook == "" {
			panic("Webhook shouldnt be empty")
		}

		reader := eventReader.NewSlackV1Events(slack.PostWebhook, webhook, logger.WithField("context", "slack-events"))
		event.GetBus().Subscribe(event.TopicMonolog, "slack-listener", reader.Read)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
		}
		event.GetBus().Close()
	},
}
