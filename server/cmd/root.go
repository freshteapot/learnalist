package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/freshteapot/learnalist-api/server/cmd/server"
	"github.com/freshteapot/learnalist-api/server/cmd/tools"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "learnalist",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.learnalist.yaml)")
	rootCmd.AddCommand(server.ServerCmd)
	rootCmd.AddCommand(tools.RootCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	if cfgFile == "" {
		fmt.Println("You need to provide a path to the config file")
		os.Exit(1)
	}

	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv() // read in environment variables that match

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Using config file:", viper.ConfigFileUsed())
}
