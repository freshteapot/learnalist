package user

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "user",
	Short: "Management of the user via the cli",
}

func init() {
	RootCmd.AddCommand(deleteUserCmd)
	RootCmd.AddCommand(findCmd)
	RootCmd.AddCommand(scratchCMD)
}
