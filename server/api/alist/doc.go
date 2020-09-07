package alist

import "errors"

var (
	ErrorListFromValid             = errors.New("validate")
	ErrorSharingNotAllowedWithFrom = errors.New("sharing-not-allowed-with-from")
)

type DatastoreAlists interface {
	GetListsByUserWithFilters(uuid string, labels string, listType string) []Alist
	GetAlist(uuid string) (Alist, error)
	GetAllListsByUser(userUUID string) []ShortInfo
	GetPublicLists() []ShortInfo
	SaveAlist(method string, aList Alist) (Alist, error)
	RemoveAlist(alistUUID string, userUUID string) error
}
