package tools

import (
	"encoding/json"
	"fmt"
	"log"
	"path/filepath"
	"strings"

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
		fmt.Println("Rebuild static sites")
		cfg := viper.Sub("tools.rebuildStaticSite")

		databaseName := cfg.GetString("sqlite.database")
		siteCacheFolder := cfg.GetString("siteCacheDirectory") // "path to site cache"
		hugoFolder := cfg.GetString("hugoDirectory")           // "path to static site builder

		hugoFolder = strings.TrimRight(hugoFolder, "/")
		siteCacheFolder = strings.TrimRight(siteCacheFolder, "/")

		// Convert paths to absolute, allowing /../x
		hugoFolder, _ = filepath.Abs(hugoFolder)
		siteCacheFolder, _ = filepath.Abs(siteCacheFolder)

		if hugoFolder == "" {
			log.Fatal("You might have forgotten to set the path to hugo directory: server.hugoDirectory")
		}

		if !utils.IsDir(hugoFolder) {
			log.Fatal(fmt.Sprintf("%s is not a directory", hugoFolder))
		}

		if siteCacheFolder == "" {
			log.Fatal("You might have forgotten to set the path to site cache directory: server.siteCacheDirectory")
		}

		if !utils.IsDir(siteCacheFolder) {
			log.Fatal(fmt.Sprintf("%s is not a directory", siteCacheFolder))
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
