package models

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/label"
	apiUser "github.com/freshteapot/learnalist-api/server/api/user"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	alist.DatastoreAlists
	apiUser.DatastoreUsers
	DatastoreUser
	DatastoreOauth2
	DatastoreLabels
}

type DatastoreOauth2 interface {
	OAuthHandler() oauth.OAuthReadWriter
}

type DatastoreLabels interface {
	Labels() label.LabelReadWriter
	RemoveUserLabel(label string, uuid string) error
}

type DatastoreUser interface {
	UserSession() user.Session
	UserFromIDP() user.UserFromIDP
	UserWithUsernameAndPassword() user.UserWithUsernameAndPassword
}
