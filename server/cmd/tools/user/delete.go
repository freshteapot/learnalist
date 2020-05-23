package user

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/spf13/cobra"
)

var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove a user from the system",
	Run: func(cmd *cobra.Command, args []string) {
		dsn, _ := cmd.Flags().GetString("dsn")
		userUUID := args[0]
		if userUUID == "" {
			fmt.Println("Can't delete an empty user")
			return
		}

		db := database.NewDB(dsn)
		m := user.NewManagement(db)
		err := m.DeleteUserFromDB(userUUID)
		fmt.Println(err)
		// Get users lists
		// Delete files from disk
		// Delete user data in hugo
		// Delete all data from sql
	},
}

func init() {
	deleteUserCmd.Flags().String("dsn", "", "Path to database")
}
