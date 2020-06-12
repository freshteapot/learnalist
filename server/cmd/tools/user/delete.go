package user

import (
	"fmt"
	"os"
	"time"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userSqlite "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
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
		dsn, _ := cmd.Flags().GetString("dsn")
		userUUID := args[0]
		if userUUID == "" {
			fmt.Println("Can't delete an empty user")
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
		//masterCron.Stop()
		hugoHelper := hugo.NewHugoHelper(hugoFolder, hugoEnvironment, hugoExternal, masterCron, logger)

		db := database.NewDB(dsn)

		// Setup access control layer.
		acl := aclStorage.NewAcl(db)
		userSession := userStorage.NewUserSession(db)
		userFromIDP := userStorage.NewUserFromIDP(db)
		userWithUsernameAndPassword := userStorage.NewUserWithUsernameAndPassword(db)
		oauthHandler := oauthStorage.NewOAuthReadWriter(db)
		dal := models.NewDAL(db, acl, userSession, userFromIDP, userWithUsernameAndPassword, oauthHandler)

		userManagement := user.NewManagement(
			userSqlite.NewSqliteManagementStorage(db),
			hugoHelper,
			event.NewInsights(logger),
		)

		err = userManagement.DeleteUser(userUUID)
		hugoHelper.WritePublicLists(dal.GetPublicLists())
		time.Sleep(1 * time.Second)
	},
}

func init() {
	deleteUserCmd.Flags().String("dsn", "", "Path to database")
}
