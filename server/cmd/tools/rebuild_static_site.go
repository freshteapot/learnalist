package tools

import (
	"encoding/json"
	"fmt"
	"os"

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

		query := `
      SELECT
    	  *
    	FROM
    		alist_kv`
		rows, _ := db.Queryx(query)

		var row models.AlistKV
		for rows.Next() {
			rows.StructScan(&row)
			aList := new(alist.Alist)
			json.Unmarshal([]byte(row.Body), &aList)
			aList.User.Uuid = row.UserUuid

			hugoHelper.Write(aList)

		}
	},
}
