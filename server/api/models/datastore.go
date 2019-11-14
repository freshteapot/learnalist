package models

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	DatastoreLabels
	DatastoreAlists
	DatastoreUsers
	DatastoreUser
	DatastoreOauth2
}

type DatastoreOauth2 interface {
	OAuthHandler() oauth.OAuthReadWriter
}

type DatastoreUser interface {
	UserSession() user.Session
	UserFromIDP() user.UserFromIDP
	UserWithUsernameAndPassword() user.UserWithUsernameAndPassword
}

type DatastoreLabels interface {
	// Labels
	PostUserLabel(label *UserLabel) (int, error)
	RemoveUserLabel(label string, uuid string) error
	PostAlistLabel(label *AlistLabel) (int, error)
}

type DatastoreAlists interface {
	// Lists
	GetUserLabels(uuid string) ([]string, error)
	GetListsByUserWithFilters(uuid string, labels string, listType string) []*alist.Alist
	GetAlist(uuid string) (*alist.Alist, error)
	//PostAlist(uuid string, aList alist.Alist) error
	SaveAlist(method string, aList alist.Alist) (*alist.Alist, error)
	//UpdateAlist(aList alist.Alist) error
	RemoveAlist(alist_uuid string, user_uuid string) error
}

type DatastoreUsers interface {
	// User
	UserExists(userUUID string) bool
}
