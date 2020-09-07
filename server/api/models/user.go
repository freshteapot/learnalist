package models

func (dal *DAL) UserExists(userUUID string) bool {
	return dal.user.UserExists(userUUID)
}
