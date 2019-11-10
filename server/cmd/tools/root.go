package tools

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tools",
	Short: "tools to make learnalist cli amazing ;)",
}

func init() {
	RootCmd.AddCommand(rebuildStaticSiteCmd)
	RootCmd.AddCommand(integrationTestsCmd)
}
