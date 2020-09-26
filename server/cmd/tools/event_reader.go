package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
)

var eventReaderCMD = &cobra.Command{
	Use:   "event-reader",
	Short: "Read events via nats",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		logger.Info("Read events")

		natsServer := "test-cluster"
		natsClientID := args[0]
		bus := event.NewNatBus(natsServer, natsClientID)
		event.SetBus(bus)
		event.GetBus().Subscribe(event.TopicMonolog, readEventLog)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
		}
		// Not great
		bus.Close(event.TopicMonolog)
	},
}

func readEventLog(entry event.Eventlog) {
	switch entry.Kind {
	case event.ApiUserLogin:
		logger := logging.GetLogger()

		userUUID := entry.Data.(string)
		logger.WithFields(logrus.Fields{
			"user_uuid": userUUID,
			"kind":      entry.Kind,
		}).Info(entry.Kind)

		fmt.Printf("%s: user %s logged in\n", entry.Kind, userUUID)
	case event.ApiUserDelete:
		logger := logging.GetLogger()

		userUUID := entry.Data.(string)
		logger.WithFields(logrus.Fields{
			"user_uuid": userUUID,
			"kind":      entry.Kind,
		}).Info(entry.Kind)
		fmt.Printf("%s: user %s should be deleted\n", entry.Kind, userUUID)
	default:
		b, _ := json.Marshal(entry)
		fmt.Println(string(b))
	}
}
