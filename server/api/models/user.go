package models

import (
	"fmt"
	"log"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
)

func (dal *DAL) UserExists(uuid string) bool {
	var id int
	query := "SELECT 1 FROM user WHERE uuid = ?"
	err := dal.Db.Get(&id, query, uuid)
	if err != nil {
		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "UserExists"))
		log.Println(err)
	}

	if id == 1 {
		return true
	}
	return false
}
