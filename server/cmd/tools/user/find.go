package user

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/spf13/cobra"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find a user based on a username or email",
	Run: func(cmd *cobra.Command, args []string) {
		dsn, _ := cmd.Flags().GetString("dsn")
		search := args[0]
		if search == "" {
			fmt.Println("Nothing to search for, means nothing to find")
			return
		}
		db := database.NewDB(dsn)
		m := user.NewManagement(db)
		userUUIDs, err := m.FindUserUUID(search)

		if err != nil {
			fmt.Println("Something went wrong")
			fmt.Println(err)
			// Printing this, as it might contain 2 results
			fmt.Println(userUUIDs)
			return
		}

		b, _ := json.Marshal(userUUIDs)
		fmt.Println(utils.PrettyPrintJSON(b))
	},
}

func init() {
	findCmd.Flags().String("dsn", "", "Path to database")
}
