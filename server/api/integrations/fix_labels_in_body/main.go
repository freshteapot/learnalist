package main

import (
	"flag"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
)

var dal *models.DAL

func setUp(databaseName string) *models.DAL {
	db := database.NewDB(databaseName)
	acl := aclStorage.NewAcl(db)
	_dal := models.NewDAL(db, acl)
	return _dal
}

func getUsers() []string {
	uuids := make([]string, 0)
	query := "SELECT uuid FROM user"
	dal.Db.Select(&uuids, query)
	return uuids
}

func main() {
	var err error
	database := flag.String("database", "/tmp/api.db", "The database.")
	flag.Parse()
	fmt.Println(`
This will:
- remove labels that are in the body of alist_kv, but not in the label tables.
- make sure that the lists without a labels attribute in info, will get an empty one.`)

	dal = setUp(*database)

	// Get all the lists
	uuids := getUsers()

	for _, uuid := range uuids {
		fmt.Println(uuid)
		lists := dal.GetListsByUserWithFilters(uuid, "", "")
		labels, _ := dal.Labels().GetUserLabels(uuid)
		// What labels are missing already
		for _, aList := range lists {
			cleaned := make([]string, 0)
			for _, label := range aList.Info.Labels {
				if utils.StringArrayContains(labels, label) {
					cleaned = append(cleaned, label)
				}
			}
			fmt.Println("Title is " + aList.Info.Title)
			fmt.Println("Uuid is " + aList.Uuid)
			fmt.Println(cleaned)
			aList.Info.Labels = cleaned
			err = dal.SaveLabelsForAlist(*aList)
			fmt.Println(err)
			_, err = dal.SaveAlist("put", *aList)
			fmt.Println(err)
		}
	}
}
