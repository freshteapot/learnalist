package tools

import (
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/fix"
	"github.com/spf13/cobra"
)

var fixInteractV1Cmd = &cobra.Command{
	Use:   "fix-interact-v1",
	Short: "Fix lists with interact object to use int not string",
	Run: func(cmd *cobra.Command, args []string) {
		dsn, _ := cmd.Flags().GetString("dsn")
		// 1) Find lists that have interact object
		// 2) Update lists of type v1 with slideshow and totalrecall
		// 3) trigger rebuild of sites?

		db := database.NewDB(dsn)

		fixup := fix.NewInteractV1(db)
		fixup.ChangeFromStringToInt()
		fixup.RemoveInteractFromNonV1()
	},
}

func init() {
	fixInteractV1Cmd.Flags().String("dsn", "", "Path to database")
}
