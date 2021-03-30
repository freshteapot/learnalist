package user

import (
	"fmt"
	"os"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/event/staticsite"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/user"

	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var deleteUserCmd = &cobra.Command{
	Use:   "delete",
	Short: "Remove a user from the system",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		os.Setenv("EVENTS_STAN_CLIENT_ID", "tools-user-mangement")
		event.SetupEventBus(logger.WithField("context", "tools-user-delete"))

		dsn := viper.GetString("server.sqlite.database")

		userUUID := args[0]
		if userUUID == "" {
			fmt.Println("Can't delete an empty user")
			return
		}

		db := database.NewDB(dsn)

		userManagement := user.NewManagement(
			userStorage.NewSqliteManagementStorage(db),
			staticsite.NewSiteManagementViaEvents(),
			event.NewInsights(logger),
		)

		err := userManagement.DeleteUser(userUUID)
		if err != nil {
			if err != utils.ErrNotFound {
				fmt.Println("Issue deleting")
				fmt.Println(err)
				return
			}
			// What can possibly go wrong if we send it thru the system?
			fmt.Println("user not found")
			//return
		}

		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: event.CMDUserDelete,
			UUID: userUUID,
		})
	},
}
