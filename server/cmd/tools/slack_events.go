package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
)

var slackEventsCMD = &cobra.Command{
	Use:   "slack-events",
	Short: "Read events and write to slack",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		logger.Info("Read events")

		webhook := viper.GetString("server.events.slack.webhook")
		if webhook == "" {
			panic("Webhook shouldnt be empty")
		}

		natsServer := viper.GetString("server.events.nats.server")
		stanClusterID := viper.GetString("server.events.stan.clusterID")
		stanClientID := viper.GetString("server.events.stan.clientID")
		fmt.Println(stanClientID)
		nats, err := nats.Connect(natsServer)
		if err != nil {
			panic(err)
		}

		event.SetBus(event.NewNatBus(stanClusterID, stanClientID, nats))

		reader := NewSlackEvents(webhook, logger.WithField("context", "slack-events"))
		event.GetBus().Subscribe(event.TopicMonolog, reader.Read)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
		}
		// Not great
		event.GetBus().Close(event.TopicMonolog)
	},
}

type SlackEvents struct {
	webhook    string
	logContext logrus.FieldLogger
}

func NewSlackEvents(webhook string, logContext logrus.FieldLogger) SlackEvents {
	return SlackEvents{
		webhook:    webhook,
		logContext: logContext,
	}
}

func (s SlackEvents) Read(entry event.Eventlog) {
	var msg slack.WebhookMessage

	switch entry.Kind {
	case event.ApiUserLogin:

		userUUID := entry.Data.(string)
		s.logContext.WithFields(logrus.Fields{
			"user_uuid": userUUID,
			"kind":      entry.Kind,
		}).Info(entry.Kind)

		msg.Text = fmt.Sprintf("%s: user %s logged in\n", entry.Kind, userUUID)
	case event.ApiUserDelete:
		userUUID := entry.Data.(string)
		s.logContext.WithFields(logrus.Fields{
			"user_uuid": userUUID,
			"kind":      entry.Kind,
		}).Info(entry.Kind)
		msg.Text = fmt.Sprintf("%s: user %s should be deleted\n", entry.Kind, userUUID)
	case event.ApiListSaved:
		b, _ := json.Marshal(entry.Data)
		var listEvent event.EventList
		json.Unmarshal(b, &listEvent)
		msg.Text = fmt.Sprintf(`list:%s (%s) %s by user:%s`, listEvent.UUID, listEvent.Data.Info.SharedWith, listEvent.Action, listEvent.UserUUID)
	case event.ApiListDelete:
		b, _ := json.Marshal(entry.Data)
		var listEvent event.EventList
		json.Unmarshal(b, &listEvent)
		msg.Text = fmt.Sprintf("list:%s deleted by user:%s", listEvent.UUID, listEvent.UserUUID)

	default:
		b, _ := json.Marshal(entry)
		fmt.Println(string(b))
		msg.Text = entry.Kind
	}

	err := slack.PostWebhook(s.webhook, &msg)
	if err != nil {
		s.logContext.Panic(err)
	}
}

func init() {
	viper.SetDefault("server.events.nats.server", "nats")
	viper.SetDefault("server.events.stan.clusterID", "stan")
	viper.SetDefault("server.events.stan.clientID", "")

	viper.BindEnv("server.events.nats.server", "EVENTS_NATS_SERVER")
	viper.BindEnv("server.events.stan.clusterID", "EVENTS_STAN_CLUSERTID")
	viper.BindEnv("server.events.stan.clientID", "EVENTS_STAN_CLIENTID")

	viper.SetDefault("server.events.slack.webhook", "")
	viper.BindEnv("server.events.slack.webhook", "EVENTS_SLACK_WEBHOOK")

}
