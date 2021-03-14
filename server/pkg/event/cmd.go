package event

import (
	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func SetDefaultSettingsForCMD() {
	viper.SetDefault("server.events.via", "nats")
	viper.BindEnv("server.events.via", "EVENTS_VIA")

	viper.SetDefault("server.events.nats.server", "nats")
	viper.SetDefault("server.events.stan.clusterID", "stan")
	viper.SetDefault("server.events.stan.clientID", "")

	viper.BindEnv("server.events.nats.server", "EVENTS_NATS_SERVER")
	viper.BindEnv("server.events.stan.clusterID", "EVENTS_STAN_CLUSTER_ID")
	viper.BindEnv("server.events.stan.clientID", "EVENTS_STAN_CLIENT_ID")
}

func SetupEventBus(logContext logrus.FieldLogger) {
	// This now works for the "application"
	eventsVia := viper.GetString("server.events.via")

	switch eventsVia {
	case "nats":
		natsServer := viper.GetString("server.events.nats.server")
		stanClusterID := viper.GetString("server.events.stan.clusterID")
		stanClientID := viper.GetString("server.events.stan.clientID")
		opts := []nats.Option{nats.Name("lal-go-server")}
		nc, err := nats.Connect(natsServer, opts...)

		if err != nil {
			panic(err)
		}

		SetBus(NewNatsBus(stanClusterID, stanClientID, nc, logContext))
	default:
		logContext.Fatal("server.events.via is not valid, memory or nats")
	}

	GetBus().Start(TopicMonolog)
	GetBus().Start(TopicStaticSite)
}
