package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/cron"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func main() {
	databaseName := flag.String("database", "/tmp/api.db", "The database.")
	siteCacheFolder := flag.String("site-cache-dir", "", "path to site cache")
	hugoFolder := flag.String("hugo-dir", "", "path to static site builder")
	flag.Parse()

	*hugoFolder = strings.TrimRight(*hugoFolder, "/")
	*siteCacheFolder = strings.TrimRight(*siteCacheFolder, "/")

	if *hugoFolder == "" {
		log.Fatal("Will need the path to site builder directory, add -hugo-dir=XXX")
	}

	if !utils.IsDir(*hugoFolder) {
		log.Fatal(fmt.Sprintf("%s is not a directory", *hugoFolder))
	}

	if *siteCacheFolder == "" {
		log.Fatal("Will need the path to site cache directory, add -site-cache-dir=XXX")
	}

	if !utils.IsDir(*siteCacheFolder) {
		log.Fatal(fmt.Sprintf("%s is not a directory", *siteCacheFolder))
	}
	// Convert paths to absolute, allowing /../x
	*hugoFolder, _ = filepath.Abs(*hugoFolder)
	*siteCacheFolder, _ = filepath.Abs(*siteCacheFolder)

	flag.Parse()

	db := database.NewDB(*databaseName)
	masterCron := cron.NewCron()
	hugoHelper := hugo.NewHugoHelper(*hugoFolder, masterCron, *siteCacheFolder)

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
}
