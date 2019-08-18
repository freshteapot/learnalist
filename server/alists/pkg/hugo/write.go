package hugo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

func (h HugoHelper) Write(aList *alist.Alist) {
	uuid := aList.Uuid
	content, _ := json.Marshal(aList)
	path := fmt.Sprintf("%s/%s.json", h.DataDirectory, uuid)
	err := ioutil.WriteFile(path, content, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
