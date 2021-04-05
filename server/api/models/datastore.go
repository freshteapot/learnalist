package models

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/label"
)

// Datastore allowing us to build an abstraction layer
type Datastore interface {
	alist.DatastoreAlists
	DatastoreLabels
}

type DatastoreLabels interface {
	Labels() label.LabelReadWriter
	RemoveUserLabel(label string, uuid string) error
}
