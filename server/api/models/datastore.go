package models

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	DatastoreAlists
	DatastoreUsers
	DatastoreUser
	DatastoreOauth2
	DatastoreLabels
}

// TODO I wonder if we need this
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

type DatastoreAlists interface {
	// Lists
	GetListsByUserWithFilters(uuid string, labels string, listType string) []alist.Alist
	GetAlist(uuid string) (alist.Alist, error)
	GetAllListsByUser(userUUID string) []alist.ShortInfo
	GetPublicLists() []alist.ShortInfo
	//PostAlist(uuid string, aList alist.Alist) error
	SaveAlist(method string, aList alist.Alist) (alist.Alist, error)
	//UpdateAlist(aList alist.Alist) error
	RemoveAlist(alist_uuid string, user_uuid string) error
}

type DatastoreUsers interface {
	// User
	UserExists(userUUID string) bool
}
