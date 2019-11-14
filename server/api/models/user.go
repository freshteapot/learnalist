package models

func (dal *DAL) UserExists(userUUID string) bool {
	var id int
	query := `
SELECT 1 FROM user WHERE uuid=?
UNION
SELECT 1 FROM user_from_idp WHERE user_uuid=?
`
	dal.Db.Get(&id, query, userUUID, userUUID)
	if id != 1 {
		return false
	}
	return true
}
