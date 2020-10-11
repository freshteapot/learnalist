package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

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
		event.SetDefaultSettingsForCMD()
		event.SetupEventBus(logger.WithField("context", "event-bus-setup"))
		event.GetBus().Subscribe("read-event", readEventLog)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
		}
		event.GetBus().Close()
	},
}

func readEventLog(entry event.Eventlog) {
	b, _ := json.Marshal(entry)
	fmt.Println(string(b))
}
