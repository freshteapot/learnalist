package tools

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/database"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/logging"
)

var rebuildStaticSiteCmd = &cobra.Command{
	Use:   "rebuild-static-site",
	Short: "Rebuild the static site based on all lists in the database",
	Run: func(cmd *cobra.Command, args []string) {
		logger := logging.GetLogger()
		event.SetDefaultSettingsForCMD()
		event.SetupEventBus(logger.WithField("context", "tools-rebuild-static-site"))

		databaseName := viper.GetString("server.sqlite.database")
		// "path to static site builder

		db := database.NewDB(databaseName)

		makeLists(db)
		makeUserLists(db)
		makePublicLists(db)

		makeChallenges(db)
		time.Sleep(2 * time.Second)
	},
}

func makeLists(db *sqlx.DB) {
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

		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistList,
			UUID:   aList.Uuid,
			Data:   *aList,
			Action: event.ActionUpdated,
		})
	}
}

func makeUserLists(db *sqlx.DB) {
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

		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			Kind:   event.ChangesetAlistUser,
			UUID:   userUUID,
			Data:   lists,
			Action: event.ActionUpdated,
		})
	}
}

func makePublicLists(db *sqlx.DB) {
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

	event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
		Kind:   event.ChangesetAlistPublic,
		Data:   lists,
		Action: event.ActionUpdated,
	})
}

func makeChallenges(db *sqlx.DB) {
	var challenges []string
	err := db.Select(&challenges, `SELECT DISTINCT(uuid) FROM challenge`)
	if err != nil {
		fmt.Println(err)
		panic("...")
	}

	challengeRepo := challenge.NewSqliteRepository(db)

	for _, challengeUUID := range challenges {
		challenge, _ := challengeRepo.Get(challengeUUID)

		event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
			UUID:   challenge.UUID,
			Kind:   event.ChangesetChallenge,
			Data:   challenge,
			Action: event.ActionUpdated,
		})
	}
}
