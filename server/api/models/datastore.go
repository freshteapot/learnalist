package models

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	DatastoreLabels
	DatastoreAlists
	DatastoreUsers
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
	InsertNewUser(loginUser authenticate.LoginUser) (*uuid.User, error)
	GetUserByCredentials(loginUser authenticate.LoginUser) (*uuid.User, error)
	UserExists(userUUID string) bool
}

type DatastoreUserSessions interface {
	Get() (string, error)
	Insert() (string, error)
}
