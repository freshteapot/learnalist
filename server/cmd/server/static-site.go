package server

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/staticsite/hugo"

	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

var StaticSiteCMD = &cobra.Command{
	Use:   "static-site",
	Short: "Run the static-site",
	Long: `
EVENTS_VIA="nats" \
EVENTS_STAN_CLIENT_ID=static-site \
EVENTS_STAN_CLUSTER_ID=test-cluster \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
static-site
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		logContext := logger.WithField("context", "static-site")

		event.SetDefaultSettingsForCMD()
		viper.Set("server.events.via", "nats")

		viper.SetDefault("topic", event.TopicStaticSite)
		event.SetupEventBus(logContext)

		viper.SetDefault("topic", event.TopicStaticSite)
		viper.BindEnv("topic", "TOPIC")
		subjectRead := viper.GetString("topic")

		// Static site
		hugoFolder, err := utils.CmdParsePathToFolder("hugo.directory", viper.GetString("hugo.directory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		hugoEnvironment := viper.GetString("hugo.environment")
		if hugoEnvironment == "" {
			fmt.Println("hugo.environment is missing")
			os.Exit(1)
		}

		hugoHelper := hugo.NewHugoHelper(hugoFolder, hugoEnvironment, logContext)

		sc := event.GetBus().(*event.NatsBus).Connection()

		subscribers := make([]event.NatsSubscriber, 0)
		subscribers = append(subscribers, hugoHelper)

		for _, subscriber := range subscribers {
			err := subscriber.Subscribe(subjectRead, sc)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to start subscriber in remind manager")
			}
		}

		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt)

		select {
		case <-signals:
			break
		}

		for _, subscriber := range subscribers {
			subscriber.Close()
		}

		event.GetBus().Close()
	},
}
