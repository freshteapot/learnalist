package natsutils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
)

var writeCMD = &cobra.Command{
	Use:   "write",
	Short: "write event to nats",
	Long: `

	TOPIC=lal.monolog \
	EVENTS_STAN_CLIENT_ID=nats-reader \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run main.go --config=../config/dev.config.yaml \
	tools natsutils write
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		viper.SetDefault("topic", "lal.monolog")
		viper.BindEnv("topic", "TOPIC")

		topic := viper.GetString("topic")
		sc := setupStan()
		reader := bufio.NewReader(os.Stdin)

		for {
			switch line, err := reader.ReadString('\n'); err {
			case nil:
				line = strings.Trim(line, "\n")

				if err := sc.Publish(topic, []byte(line)); err != nil {
					logger.Fatal(err)
				}
			case io.EOF:
				return
			default:
				fmt.Println(err)
				os.Exit(1)
				return
			}
		}
	},
}

func setupStan() stan.Conn {
	logger := logging.GetLogger()
	natsServer := viper.GetString("server.events.nats.server")
	clusterID := viper.GetString("server.events.stan.clusterID")
	clientID := viper.GetString("server.events.stan.clientID")
	opts := []nats.Option{nats.Name("lal-go-server")}
	nc, err := nats.Connect(natsServer, opts...)

	if err != nil {
		panic(err)
	}

	logContext := logger.WithFields(logrus.Fields{
		"context":    "monolog",
		"cluster_id": clusterID,
		"client_id":  clientID,
	})

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
	return sc
}
