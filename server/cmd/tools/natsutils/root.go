package natsutils

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "natsutils",
	Short: "nats utils",
}

func init() {
	RootCmd.AddCommand(readCMD)
}
