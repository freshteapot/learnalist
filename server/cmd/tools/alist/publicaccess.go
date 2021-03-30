package alist

import (
	"fmt"
	"net/http"
	"os"

	alistStorage "github.com/freshteapot/learnalist-api/server/api/alist/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/event/staticsite"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var publicAccessCMD = &cobra.Command{
	Use:   "public-access",
	Short: "Grant or revoke user access to writing public lists",
	Args:  cobra.ExactArgs(1),
	Long: `
Grant user access to write public lists
Revoke user access to write public lists and set the users public lists to private

Example:

go run main.go --config=../config/dev.config.yaml \
tools list public-access chris --access=revoke
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		os.Setenv("EVENTS_STAN_CLIENT_ID", "tools-alist")
		event.SetupEventBus(logger.WithField("context", "tools-list-public-access"))

		dsn := viper.GetString("server.sqlite.database")
		db := database.NewDB(dsn)
		aclRepo := aclStorage.NewAcl(db)

		storageAlist := alistStorage.NewAlist(db, logger)

		userManagement := user.NewManagement(
			userStorage.NewSqliteManagementStorage(db),
			staticsite.NewSiteManagementViaEvents(),
			event.NewInsights(logger),
		)

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
		allowed := []string{"grant", "revoke"}
		if !utils.StringArrayContains(allowed, accessType) {
			fmt.Println("Access can be only grant or revoke")
			return
		}

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

		if accessType == "grant" {
			return
		}

		logContext := logger.WithField("sub-context", "revoke")
		shortLists := storageAlist.GetAllListsByUser(userUUID)
		changed := make([]string, 0)
		for _, short := range shortLists {
			aList, err := storageAlist.GetAlist(short.UUID)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to get list from storage")
			}

			if aList.Info.SharedWith != keys.SharedWithPublic {
				continue
			}
			// Change to private
			aList.Info.SharedWith = keys.NotShared

			_, err = storageAlist.SaveAlist(http.MethodPut, aList)
			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to update list to storage")
			}

			err = aclRepo.MakeListPrivate(aList.Uuid, userUUID)

			if err != nil {
				logContext.WithFields(logrus.Fields{
					"error": err,
				}).Fatal("Failed to set list to private")
			}

			event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
				Kind: event.ApiListSaved,
				Data: event.EventList{
					UUID:     aList.Uuid,
					UserUUID: userUUID,
					Action:   event.ActionUpdated,
					Data:     aList,
				},
			})
			changed = append(changed, aList.Uuid)
		}

		logContext.WithFields(logrus.Fields{
			"total": len(changed),
			"lists": changed,
		}).Info("lists revoked")
	},
}

func init() {
	publicAccessCMD.Flags().Bool("current", false, "Show current for user")
	publicAccessCMD.Flags().String("access", "", "revoke / grant")
}
