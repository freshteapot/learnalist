package tools

import (
	"github.com/freshteapot/learnalist-api/server/cmd/tools/api"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/docs"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/user"
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "tools",
	Short: "tools to make learnalist cli amazing ;)",
}

func init() {
	RootCmd.AddCommand(rebuildStaticSiteCmd)
	RootCmd.AddCommand(integrationTestsCmd)
	RootCmd.AddCommand(stubSQLFileCmd)
	RootCmd.AddCommand(fixInteractV1Cmd)
	RootCmd.AddCommand(user.RootCmd)
	RootCmd.AddCommand(api.RootCmd)
	RootCmd.AddCommand(docs.RootCmd)
}
