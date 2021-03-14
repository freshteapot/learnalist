package models

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	alist.DatastoreAlists
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
