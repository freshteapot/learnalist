package remind

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "remind",
	Short: "remind",
}

func init() {
	RootCmd.AddCommand(dailyCMD)
}
