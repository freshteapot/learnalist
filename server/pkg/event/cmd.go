package event

import "github.com/spf13/viper"

func SetDefaultSettingsForCMD() {
	viper.SetDefault("server.events.nats.server", "nats")
	viper.SetDefault("server.events.stan.clusterID", "stan")
	viper.SetDefault("server.events.stan.clientID", "")

	viper.BindEnv("server.events.nats.server", "EVENTS_NATS_SERVER")
	viper.BindEnv("server.events.stan.clusterID", "EVENTS_STAN_CLUSTER_ID")
	viper.BindEnv("server.events.stan.clientID", "EVENTS_STAN_CLIENT_ID")
}
