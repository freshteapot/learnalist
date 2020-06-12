package hugo

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
)

func (h HugoHelper) WriteList(aList alist.Alist) {
	h.AlistWriter.Data(aList)
	h.AlistWriter.Content(aList)
	h.RegisterCronJob()
}

// WriteListsByUser
func (h HugoHelper) WriteListsByUser(userUUID string, lists []alist.ShortInfo) {
	h.AlistsByUserWriter.Data(userUUID, lists)
	h.AlistsByUserWriter.Content(userUUID)
	h.RegisterCronJob()
}

func (h HugoHelper) WritePublicLists(lists []alist.ShortInfo) {
	h.PublicListsWriter.Data(lists)
	h.RegisterCronJob()
}
