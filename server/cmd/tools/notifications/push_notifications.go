package notifications

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	firebase "firebase.google.com/go/v4"
	messaging "firebase.google.com/go/v4/messaging"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"google.golang.org/api/option"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
)

var pushNotificationsCMD = &cobra.Command{
	Use:   "push-notifications",
	Short: "Read events via nats",
	Long: `

ssh $SSH_SERVER -L 4222:127.0.0.1:4222 -N &
ssh $SSH_SERVER sudo kubectl port-forward deployment/stan01 4222:4222 &

TOPIC=notifications \
EVENTS_STAN_CLIENT_ID=push-notifications \
EVENTS_STAN_CLUSTER_ID=test-cluster \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml tools notifications push-notifications

	`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		// Leaving empty, but should be notifications
		viper.SetDefault("topic", "")
		viper.BindEnv("topic", "TOPIC")
		topic := viper.GetString("topic")
		if topic == "" {
			logger.Fatal("topic needs setting")
		}

		natsServer := viper.GetString("server.events.nats.server")
		clusterID := viper.GetString("server.events.stan.clusterID")
		clientID := viper.GetString("server.events.stan.clientID")
		pathToCredentials := viper.GetString("server.fcm.credentials")
		opts := []nats.Option{nats.Name("lal-go-server")}
		nc, err := nats.Connect(natsServer, opts...)

		if err != nil {
			panic(err)
		}

		logContext := logger.WithFields(logrus.Fields{
			"context":    "push-notifications",
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

		opt := option.WithCredentialsFile(pathToCredentials)
		app, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Fatalf("error initializing app: %v\n", err)
		}

		ctx := context.Background()
		client, err := app.Messaging(ctx)
		if err != nil {
			log.Fatalf("error getting Messaging client: %v\n", err)
		}

		handle := func(msg *stan.Msg) {
			var moment event.Eventlog
			json.Unmarshal(msg.Data, &moment)

			if moment.Kind != event.KindPushNotification {
				return
			}

			var message *messaging.Message
			b, _ := json.Marshal(moment.Data)
			json.Unmarshal(b, &message)

			// Send a message to the device corresponding to the provided
			// registration token.
			response, err := client.Send(ctx, message)
			if err != nil {
				if err.Error() == "The registration token is not a valid FCM registration token" {
					event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
						Kind: event.MobileDeviceRemove,
						Data: message.Token,
					})

					logContext.WithFields(logrus.Fields{
						"event": "invalid",
						"token": message.Token,
					}).Error("bad token")
					return
				}

				if err.Error() == "Requested entity was not found." {
					event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
						Kind: event.MobileDeviceRemove,
						Data: message.Token,
					})

					// Poor mans option
					// cat events.ndjson | jq -r 'select(.event=="stale") | "DELETE FROM mobile_device WHERE token=\"\(.token)\";"'
					logContext.WithFields(logrus.Fields{
						"event": "stale",
						"token": message.Token,
					}).Error("bad token")
					return
				}
				log.Fatalln(err)
			}
			logContext.WithField("response", response).Info("success")
		}

		durableName := "challenges.pushNotifications"
		subscription, _ := sc.Subscribe(
			topic,
			handle,
			stan.DurableName(durableName),
			stan.MaxInflight(1),
		)

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
			break
		}

		subscription.Close()
		sc.Close()
	},
}
