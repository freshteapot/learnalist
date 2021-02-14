package tools

import (
	"github.com/freshteapot/learnalist-api/server/cmd/tools/api"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/challenges"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/docs"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/experimental"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/natsutils"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/notifications"
	"github.com/freshteapot/learnalist-api/server/cmd/tools/remind"
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
	RootCmd.AddCommand(fixPlankV1Cmd)
	RootCmd.AddCommand(user.RootCmd)
	RootCmd.AddCommand(api.RootCmd)
	RootCmd.AddCommand(natsutils.RootCmd)
	RootCmd.AddCommand(challenges.RootCmd)
	RootCmd.AddCommand(notifications.RootCmd)
	RootCmd.AddCommand(docs.RootCmd)
	RootCmd.AddCommand(remind.RootCmd)
	RootCmd.AddCommand(eventReaderCMD)
	RootCmd.AddCommand(slackEventsCMD)
	RootCmd.AddCommand(experimental.RootCmd)
}
