package docs

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "docs",
	Short: "Documentation builders",
}

func init() {
	RootCmd.AddCommand(apiOverviewCMD)
}
