package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

var rebuildStaticSiteCmd = &cobra.Command{
	Use:   "rebuild-static-site",
	Short: "Rebuild the static site based on all lists in the database",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		skipPublishing, _ := cmd.Flags().GetBool("skip-publishing")
		databaseName := viper.GetString("tools.rebuildStaticSite.sqlite.database")
		// "path to static site builder
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

		db := database.NewDB(databaseName)
		masterCron := cron.NewCron()
		if skipPublishing {
			masterCron.Stop()
		}
		hugoHelper := hugo.NewHugoHelper(hugoFolder, hugoEnvironment, hugoExternal, masterCron, logger)

		makeLists(db, hugoHelper)
		makeUserLists(db, hugoHelper)
		makePublicLists(db, hugoHelper)

		time.Sleep(5 * time.Second)
	},
}

func init() {
	rebuildStaticSiteCmd.Flags().Bool("skip-publishing", false, "Skip the actual building of the pages")
}

func makeLists(db *sqlx.DB, helper hugo.HugoSiteBuilder) {
	query := `
SELECT
	*
FROM
	alist_kv`
	rows, _ := db.Queryx(query)

	for rows.Next() {
		var row models.AlistKV
		rows.StructScan(&row)
		aList := new(alist.Alist)
		json.Unmarshal([]byte(row.Body), &aList)
		aList.User.Uuid = row.UserUuid

		helper.WriteList(*aList)
	}
}

func makeUserLists(db *sqlx.DB, helper hugo.HugoSiteBuilder) {
	var users []string
	err := db.Select(&users, `SELECT DISTINCT(user_uuid) FROM alist_kv`)
	if err != nil {
		fmt.Println(err)
		panic("...")
	}

	for _, userUUID := range users {
		var lists []alist.ShortInfo
		query := `
SELECT
	json_extract(body, '$.info.title') AS title,
	uuid
FROM
	alist_kv
WHERE
	user_uuid=?`

		err := db.Select(&lists, query, userUUID)
		if err != nil {
			fmt.Println(err)
			panic("...")
		}

		helper.WriteListsByUser(userUUID, lists)
	}
}

func makePublicLists(db *sqlx.DB, helper hugo.HugoSiteBuilder) {
	query := `
SELECT
	uuid,
	title
FROM (
SELECT
	json_extract(body, '$.info.title') AS title,
	IFNULL(json_extract(body, '$.info.shared_with'), "private") AS shared_with,
	uuid
FROM
	alist_kv
) as temp
WHERE shared_with="public";
`
	var lists []alist.ShortInfo
	err := db.Select(&lists, query)
	if err != nil {
		fmt.Println(err)
		panic("Failed to make public lists")
	}
	helper.WritePublicLists(lists)
}
