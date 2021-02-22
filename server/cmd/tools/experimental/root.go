package experimental

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "experimental",
	Short: "experimental",
}

func init() {
	RootCmd.AddCommand(extractEventsCMD)
}
