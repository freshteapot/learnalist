package models

import "github.com/freshteapot/learnalist-api/server/pkg/user"

func (dal *DAL) UserSession() user.Session {
	return dal.userSession
}

func (dal *DAL) UserFromIDP() user.UserFromIDP {
	return dal.userFromIDP
}
