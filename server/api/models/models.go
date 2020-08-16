package models

import (
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/jmoiron/sqlx"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db                          *sqlx.DB
	Acl                         acl.Acl
	userSession                 user.Session
	userFromIDP                 user.UserFromIDP
	userWithUsernameAndPassword user.UserWithUsernameAndPassword
	oauthHandler                oauth.OAuthReadWriter
	labels                      label.LabelReadWriter
}

func NewDAL(db *sqlx.DB, acl acl.Acl, labels label.LabelReadWriter, userSession user.Session, userFromIDP user.UserFromIDP, userWithUsernameAndPassword user.UserWithUsernameAndPassword, oauthHandler oauth.OAuthReadWriter) *DAL {
	dal := &DAL{
		Db:                          db,
		Acl:                         acl,
		userSession:                 userSession,
		userFromIDP:                 userFromIDP,
		userWithUsernameAndPassword: userWithUsernameAndPassword,
		oauthHandler:                oauthHandler,
		labels:                      labels,
	}
	return dal
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
