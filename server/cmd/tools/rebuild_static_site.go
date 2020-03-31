package tools

import (
	"encoding/json"
	"fmt"
	"os"

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
		makeUserLists(db, hugoHelper)
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
		var lists []string
		err := db.Select(&lists, `SELECT body FROM alist_kv WHERE user_uuid=?`, userUUID)
		if err != nil {
			fmt.Println(err)
			panic("...")
		}

		aLists := make([]alist.Alist, len(lists))
		for index, raw := range lists {
			var aList alist.Alist
			aList.UnmarshalJSON([]byte(raw))
			aLists[index] = aList
		}

		helper.WriteListsByUser(userUUID, aLists)
	}

}
