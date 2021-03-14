package hugo

import (
	"encoding/json"
	"fmt"

	"github.com/freshteapot/learnalist-api/server/api/alist"
)

func NewHugoPublicListsWriter(dataDirectory string, publishDirectory string, writer FileWriter) HugoPublicListsWriter {
	return HugoPublicListsWriter{
		dataDirectory:    dataDirectory,
		publishDirectory: publishDirectory,
		writer:           writer,
	}
}

func (w HugoPublicListsWriter) Content() {
	// I think it might be best to not have anything here
	// Instead, define in a particular partial or layout in hugo
}

func (w HugoPublicListsWriter) Data(lists []alist.ShortInfo) {
	content, _ := json.Marshal(lists)
	path := fmt.Sprintf("%s/public_lists.json", w.dataDirectory)
	w.writer.Write(path, content)
}
