package hugo

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

// TODO will need to handle this
// Remove delete cached files based on list uui.
func (h HugoHelper) Remove(uuid string) {
	//h.deleteBuildFiles(uuid)

	files := []string{
		fmt.Sprintf("%s/alist/%s.html", h.SiteCacheFolder, uuid),
		fmt.Sprintf("%s/alist/%s.json", h.SiteCacheFolder, uuid),
	}
	h.deleteFiles(files)
}

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
	fmt.Println("WritePublicLists")
	h.PublicListsWriter.Data(lists)
	h.RegisterCronJob()
}
