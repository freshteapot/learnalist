package api

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "api",
	Short: "Management of the user via the cli",
}

func init() {
	RootCmd.AddCommand(deleteUserCmd)
}
