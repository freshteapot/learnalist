package tools

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/freshteapot/learnalist-api/server/api/database"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
)

var fixACLOwnerCMD = &cobra.Command{
	Use:   "fix-acl-owner",
	Short: "Make sure lists with no shared settings are set to private",
	Run: func(cmd *cobra.Command, args []string) {
		databaseName := viper.GetString("server.sqlite.database")
		db := database.NewDB(databaseName)

		acl := aclStorage.NewAcl(db)

		type tempListInfo struct {
			AlistUUID string `db:"uuid" json:"uuid"`
			UserUUID  string `db:"user_uuid" json:"user_uuid"`
		}

		query := `
SELECT
	uuid,
	user_uuid
FROM
	alist_kv
WHERE
	json_extract(body, '$.info.shared_with') IS NULL;
`
		var lists []tempListInfo
		db.Select(&lists, query)

		for _, info := range lists {
			fmt.Printf("Setting list:%s to private for user:%s\n", info.AlistUUID, info.UserUUID)
			// TODO actually need to resave the lists :(
			// dal.alist.SaveAlist(method, aList)
			err := acl.MakeListPrivate(info.AlistUUID, info.UserUUID)
			fmt.Println(err)
		}
	},
}
