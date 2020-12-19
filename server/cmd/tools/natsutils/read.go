package natsutils

import (
	"context"
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

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
)

var readCMD = &cobra.Command{
	Use:   "read",
	Short: "Read events via nats",
	Long: `

	kubectl port-forward svc/nats 4222:4222 &


	TOPIC=lal.monolog \
	EVENTS_STAN_CLIENT_ID=nats-reader \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run main.go --config=../config/dev.config.yaml \
	tools natsutils read
	`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		logger.Info("Read events")
		event.SetDefaultSettingsForCMD()
		viper.SetDefault("topic", "lal.monolog")
		viper.BindEnv("topic", "TOPIC")

		topic := viper.GetString("topic")
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
		defer logCloser(sc)

		ctx := context.Background()
		ReadOneByOneTilLatest(sc, topic, func(msg *stan.Msg) {
			fmt.Println(string(msg.Data))
		}, false)

		ctx, cancel := context.WithCancel(context.Background())
		latestSubscription := readLatest(ctx, sc, topic, func(msg *stan.Msg) {
			fmt.Println(string(msg.Data))
		})

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
			cancel()
		}
		latestSubscription.Unsubscribe()
	},
}

func ReadOneByOneTilLatest(sc stan.Conn, topic string, onRead func(msg *stan.Msg), unsubscribe bool) {
	// Could also use MaxInFlight with the timer to force one by one
	d := 200 * time.Millisecond
	// Initially we shall wait
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	handle := func(msg *stan.Msg) {
		ticker.Stop()
		onRead(msg)
		ticker.Reset(d)
	}

	durableName := "reader"
	subscription, _ := sc.Subscribe(
		topic,
		handle,
		stan.DurableName(durableName),
		stan.DeliverAllAvailable(),
		stan.MaxInflight(1),
	)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	select {
	case <-done:
		break
	//case t := <-ticker.C:
	case <-ticker.C:
		break
	}

	if !unsubscribe {
		subscription.Close()
		return
	}

	subscription.Unsubscribe()
	return
}

func readLatest(ctx context.Context, sc stan.Conn, topic string, onRead func(msg *stan.Msg)) stan.Subscription {
	durableName := "reader"
	subscription, _ := sc.Subscribe(
		topic,
		onRead,
		stan.DurableName(durableName),
		stan.MaxInflight(1),
	)
	return subscription
}

func logCloser(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Printf("close error: %s", err)
	}
}
