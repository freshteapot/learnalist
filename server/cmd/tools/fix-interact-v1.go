package tools

import (
	"fmt"

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
		history := fix.NewHistory(db)
		fixup := fix.NewInteractV1(db)
		exists, err := history.Exists(fixup)
		if err != nil {
			fmt.Println("Failed to check if the fix has been applied")
			fmt.Println(err)
			return
		}

		if exists {
			fmt.Println("Already ran")
			return
		}

		fixup.ChangeFromStringToInt()
		fixup.RemoveInteractFromNonV1()

		err = history.Save(fixup)
		if err != nil {
			fmt.Println("Failed to save, this is an utter mess, panic!")
		}
	},
}

func init() {
	fixInteractV1Cmd.Flags().String("dsn", "", "Path to database")
}
