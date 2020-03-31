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
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

var rebuildStaticSiteCmd = &cobra.Command{
	Use:   "rebuild-static-site",
	Short: "Rebuild the static site based on all lists in the database",
	Run: func(cmd *cobra.Command, args []string) {
		databaseName := viper.GetString("tools.rebuildStaticSite.sqlite.database")
		// "path to static site builder
		hugoFolder, err := utils.CmdParsePathToFolder("tools.rebuildStaticSite.hugoDirectory", viper.GetString("tools.rebuildStaticSite.hugoDirectory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		// "path to site cache"
		siteCacheFolder, err := utils.CmdParsePathToFolder("tools.rebuildStaticSite.siteCacheDirectory", viper.GetString("tools.rebuildStaticSite.siteCacheDirectory"))
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		db := database.NewDB(databaseName)
		masterCron := cron.NewCron()
		hugoHelper := hugo.NewHugoHelper(hugoFolder, masterCron, siteCacheFolder)

		makeLists(db, hugoHelper)
		//makeUserLists(db, hugoHelper)
		fmt.Println("sleeping")
		time.Sleep(time.Duration(5 * time.Second))
		fmt.Println("slept")
	},
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

		helper.WriteList(aList)
	}
	fmt.Println("HI")
}

func makeUserLists(db *sqlx.DB, helper hugo.HugoSiteBuilder) {
	// This will break one day
	queryUsers := `
SELECT
	uuid, user_uuid
FROM
	alist_kv`

	queryListsByUser := `
SELECT
	uuid
FROM
	alist_kv
WHERE
	user_uuid = ?
`

	rows, _ := db.Queryx(queryUsers)

	for rows.Next() {
		var row models.AlistKV
		rows.StructScan(&row)

		userUUID := row.UserUuid

		listRows, _ := db.Queryx(queryListsByUser, userUUID)
		lists := []string{}
		for listRows.Next() {
			var listRow models.AlistKV
			listRows.StructScan(&listRow)
			lists = append(lists, listRow.Uuid)
		}
		helper.WriteListsByUser(userUUID, lists)

	}

}
