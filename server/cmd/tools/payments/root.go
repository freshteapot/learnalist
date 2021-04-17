package payments

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "payments",
	Short: "payments commands",
}

func init() {
	RootCmd.AddCommand(slackCMD)
}
