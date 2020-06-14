package hugo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

func NewHugoPublicListsWriter(dataDirectory string) HugoPublicListsWriter {
	return HugoPublicListsWriter{
		dataDirectory: dataDirectory,
	}
}

func (w HugoPublicListsWriter) Content() {
	// I think it might be best to not have anything here
	// Instead, define in a particular partial or layout in hugo
}

func (w HugoPublicListsWriter) Data(lists []alist.ShortInfo) {
	content, _ := json.Marshal(lists)
	path := fmt.Sprintf("%s/public_lists.json", w.dataDirectory)
	err := ioutil.WriteFile(path, content, 0644)
	if err != nil {
		fmt.Println(err)
	}
}
