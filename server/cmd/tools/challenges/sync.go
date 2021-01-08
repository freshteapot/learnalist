package challenges

import (
	"encoding/json"
	"os"
	"os/signal"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

var syncCMD = &cobra.Command{
	Use:   "sync",
	Short: "Copy across challenge specific events",
	Long: `

	kubectl port-forward svc/nats 4222:4222 &

	TOPIC=lal.monolog \
	EVENTS_STAN_CLIENT_ID=challenges-sync \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run main.go --config=../config/dev.config.yaml \
	tools challenge sync

	`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		viper.SetDefault("topic", "lal.monolog")
		viper.BindEnv("topic", "TOPIC")
		subjectRead := viper.GetString("topic")
		subjectWrite := "challenges"

		natsServer := viper.GetString("server.events.nats.server")
		clusterID := viper.GetString("server.events.stan.clusterID")
		clientID := viper.GetString("server.events.stan.clientID")
		opts := []nats.Option{nats.Name("lal-go-server")}
		nc, err := nats.Connect(natsServer, opts...)

		if err != nil {
			panic(err)
		}

		logContext := logger.WithFields(logrus.Fields{
			"context":    "challenges-sync",
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

		allowed := []string{
			challenge.EventChallengeNewRecord,
			challenge.EventChallengeCreated,
			challenge.EventChallengeDeleted,
			challenge.EventChallengeJoined,
			challenge.EventChallengeLeft,
		}

		handle := func(msg *stan.Msg) {
			var moment event.Eventlog
			json.Unmarshal(msg.Data, &moment)

			if !utils.StringArrayContains(allowed, moment.Kind) {
				return
			}

			// Keep the time
			moment.Timestamp = msg.Timestamp
			b, _ := json.Marshal(moment)
			// The context of time is lost :(
			err := sc.Publish(subjectWrite, b)
			if err != nil {
				logContext.WithField("error", err).Fatal("error publishing")
			}
		}

		durableName := "challenges.sync"
		subscription, _ := sc.Subscribe(
			subjectRead,
			handle,
			stan.DurableName(durableName),
			stan.DeliverAllAvailable(),
			stan.MaxInflight(1),
		)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
			break
		}

		err = subscription.Close()
		if err != nil {
			logContext.WithField("error", err).Error("closing subscription")
		}

		err = sc.Close()
		if err != nil {
			logContext.WithField("error", err).Error("closing stan")
		}
	},
}
