package hugo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

func (h HugoHelper) WriteList(aList *alist.Alist) {
	uuid := aList.Uuid
	content, _ := json.Marshal(aList)
	path := fmt.Sprintf("%s/%s.json", h.DataDirectory, uuid)
	err := ioutil.WriteFile(path, content, 0644)
	if err != nil {
		fmt.Println(err)
	}

	h.RegisterCronJob()
}

// Remove delete cached files based on list uuid.
func (h HugoHelper) Remove(uuid string) {
	h.deleteBuildFiles(uuid)

	files := []string{
		fmt.Sprintf("%s/alist/%s.html", h.SiteCacheFolder, uuid),
		fmt.Sprintf("%s/alist/%s.json", h.SiteCacheFolder, uuid),
	}
	h.deleteFiles(files)
}
