package payment

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "payment",
	Short: "payment commands",
}

func init() {
	RootCmd.AddCommand(slackCMD)
}
