package tools

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alistStorage "github.com/freshteapot/learnalist-api/server/api/alist/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/database"
	labelStorage "github.com/freshteapot/learnalist-api/server/api/label/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/models"
	apiUserStorage "github.com/freshteapot/learnalist-api/server/api/user/sqlite"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/fix"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"

	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var fixPlankV1Cmd = &cobra.Command{
	Use:   "fix-plank-v1",
	Short: "Move planks from lists to plank table",
	Long: `
Step 1: fix planks to table

	go run -tags=json1 main.go --config=../config/dev.config.yaml tools fix-plank-v1

Step2: rebuild static site

	HUGO_EXTERNAL=false go run -tags=json1 main.go --config=../config/dev.config.yaml tools rebuild-static-site
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		databaseName := viper.GetString("server.sqlite.database")

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

		masterCron := cron.NewCron()
		hugoHelper := hugo.NewHugoHelper(hugoFolder, hugoEnvironment, hugoExternal, masterCron, logger)

		db := database.NewDB(databaseName)

		// Setup access control layer.
		acl := aclStorage.NewAcl(db)
		userSession := userStorage.NewUserSession(db)
		userFromIDP := userStorage.NewUserFromIDP(db)
		userWithUsernameAndPassword := userStorage.NewUserWithUsernameAndPassword(db)
		oauthHandler := oauthStorage.NewOAuthReadWriter(db)
		labels := labelStorage.NewLabel(db)
		storageAlist := alistStorage.NewAlist(db, logger)
		storageApiUser := apiUserStorage.NewUser(db)
		dal := models.NewDAL(acl, storageApiUser, storageAlist, labels, userSession, userFromIDP, userWithUsernameAndPassword, oauthHandler)

		repo := plank.NewSqliteRepository(db)
		history := fix.NewHistory(db)
		fixup := fix.NewPlankV1(db)
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

		records := fixup.GetPlankRecords()

		for _, item := range records {
			alistUUID := item.AlistUUID
			userUUID := item.UserUUID
			var rawData []interface{}
			json.Unmarshal([]byte(item.Data), &rawData)

			// Loop over the string entries
			for _, rawPlank := range rawData {
				var record plank.HttpRequestInput

				json.Unmarshal([]byte(rawPlank.(string)), &record)

				record.UUID = ""
				b, _ := json.Marshal(record)
				hash := fmt.Sprintf("%x", sha1.Sum(b))
				record.UUID = hash
				created := time.Unix(0, int64(record.BeginningTime)*int64(1000000))

				item := plank.Entry{
					UserUUID: userUUID,
					UUID:     hash,
					Body:     record,
					Created:  created.UTC(),
				}

				err := repo.SaveEntry(item)
				if err != nil {
					if err != plank.ErrEntryExists {
						fmt.Println(rawPlank)
						fmt.Println("error on save, abort.", err)
						panic("rewind...")
					}

				}
			}

			dal.RemoveAlist(alistUUID, userUUID)
			hugoHelper.DeleteList(alistUUID)
			hugoHelper.WriteListsByUser(userUUID, dal.GetAllListsByUser(userUUID))
		}

		err = history.Save(fixup)
		if err != nil {
			fmt.Println("Failed to save, this is an utter mess, panic!")
			os.Exit(1)
		}
	},
}
