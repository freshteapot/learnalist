package user

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userSqlite "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var findCmd = &cobra.Command{
	Use:   "find",
	Short: "Find a user based on a username or email",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		event.SetupEventBus(logger.WithField("context", "tools-user-find"))

		dsn := viper.GetString("server.sqlite.database")
		search := args[0]
		if search == "" {
			fmt.Println("Nothing to search for, means nothing to find")
			return
		}

		hugoFolder, err := utils.CmdParsePathToFolder("hugo.directory", viper.GetString("hugo.directory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		hugoEnvironment := viper.GetString("hugo.environment")
		if hugoEnvironment == "" {
			fmt.Println("hugo.environment is missing")
			os.Exit(1)
		}

		hugoExternal := viper.GetBool("hugo.external")
		if hugoEnvironment == "" {
			fmt.Println("hugo.external is missing")
			os.Exit(1)
		}

		masterCron := cron.NewCron()
		masterCron.Stop()
		hugoHelper := hugo.NewHugoHelper(hugoFolder, hugoEnvironment, hugoExternal, masterCron, logger)

		db := database.NewDB(dsn)
		userManagement := user.NewManagement(
			userSqlite.NewSqliteManagementStorage(db),
			hugoHelper,
			event.NewInsights(logger),
		)

		userUUIDs, err := userManagement.FindUser(search)

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
