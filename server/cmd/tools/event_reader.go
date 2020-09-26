package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/nats-io/stan.go"
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

func readEventLog(stanMsg *stan.Msg) {
	var entry event.Eventlog
	err := json.Unmarshal(stanMsg.Data, &entry)
	if err != nil {
		return
	}

	switch entry.Kind {
	case event.ApiUserLogin:
		userUUID := string(entry.Data)
		fmt.Printf("User %s logged in", userUUID)
	default:
		fmt.Println(entry.Kind)
		fmt.Println(string(entry.Data))
	}
}
