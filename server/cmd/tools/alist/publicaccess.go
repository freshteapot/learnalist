package alist

import (
	"fmt"
	"os"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/event/staticsite"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var publicAccessCMD = &cobra.Command{
	Use:   "public-access",
	Short: "Grant or revoke user access to writing public lists",
	Args:  cobra.ExactArgs(1),
	Long: `
Example:

go run main.go --config=../config/dev.config.yaml \
tools list public-access chris --access=revoke
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Grant = just set acl
		// Revoke = set acl and set all public to private
		// Lookup user lists which are public (do I have this feature?)
		// if list is public, update to private
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		os.Setenv("EVENTS_STAN_CLIENT_ID", "tools-alist")
		event.SetupEventBus(logger.WithField("context", "tools-list-public-access"))

		dsn := viper.GetString("server.sqlite.database")
		db := database.NewDB(dsn)
		aclRepo := aclStorage.NewAcl(db)

		userUUID := args[0]
		if userUUID == "" {
			fmt.Println("User UUID is missing")
			return
		}
		current, _ := cmd.Flags().GetBool("current")

		if current {
			fmt.Printf("Lookup user:%s status\n", userUUID)
			fmt.Println(aclRepo.HasUserPublicListWriteAccess(userUUID))
			return
		}

		accessType, _ := cmd.Flags().GetString("access")
		allowed := []string{"grant", "revoke", "current"}
		if !utils.StringArrayContains(allowed, accessType) {
			fmt.Println("Access can be only grant or revoke")
			return
		}

		userManagement := user.NewManagement(
			userStorage.NewSqliteManagementStorage(db),
			staticsite.NewSiteManagementViaEvents(),
			event.NewInsights(logger),
		)

		exists := userManagement.UserExists(userUUID)
		if !exists {
			fmt.Printf("Couldnt find user:%s\n", userUUID)
			os.Exit(1)
		}

		fmt.Printf("Set user:%s access to %s\n", userUUID, accessType)

		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: acl.EventPublicListAccess,
			Data: acl.EventPublicListAccessData{
				UserUUID: userUUID,
				Action:   accessType,
			},
			TriggeredBy: "cmd",
		})

		// TODO if revoke do more
		// Change public lists to private
	},
}

func init() {
	publicAccessCMD.Flags().Bool("current", false, "Show current for user")
	publicAccessCMD.Flags().String("access", "", "revoke / grant")
}
