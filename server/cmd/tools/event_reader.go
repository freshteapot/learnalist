package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

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

		natsServer := viper.GetString("server.events.nats.server")
		stanClusterID := viper.GetString("server.events.stan.clusterID")
		stanClientID := viper.GetString("server.events.stan.clientID")
		nats, err := nats.Connect(natsServer)
		if err != nil {
			panic(err)
		}

		event.SetBus(event.NewNatsBus(stanClusterID, stanClientID, nats, logger))
		event.GetBus().Start()
		event.GetBus().Subscribe("read-event", readEventLog)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
		}
		// Not great
		event.GetBus().Close()
	},
}

func readEventLog(entry event.Eventlog) {
	b, _ := json.Marshal(entry)
	fmt.Println(string(b))
}
