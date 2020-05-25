package hugo

import (
	"fmt"
)

func (h HugoHelper) DeleteList(listUUID string) error {
	contentDirectory := fmt.Sprintf(RealtivePathContentAlist, h.Cwd)
	dataDirectory := fmt.Sprintf(RealtivePathDataAlist, h.Cwd)
	publishDirectory := h.PublishDirectory

	files := []string{
		fmt.Sprintf("%s/%s.md", contentDirectory, listUUID),
		fmt.Sprintf("%s/%s.json", dataDirectory, listUUID),
		// Not sure if this one is in use
		fmt.Sprintf("%s/alist/%s.json", publishDirectory, listUUID),
		fmt.Sprintf("%s/alist/%s.html", publishDirectory, listUUID),
	}

	h.deleteFiles(files)
	// TODO do I need to return anything?
	return nil
}

func (h HugoHelper) DeleteUser(userUUID string) error {
	contentDirectory := fmt.Sprintf(RealtivePathContentAlistsByUser, h.Cwd)
	dataDirectory := fmt.Sprintf(RealtivePathDataAlistsByUser, h.Cwd)
	publishDirectory := h.PublishDirectory

	files := []string{
		fmt.Sprintf("%s/%s.md", contentDirectory, userUUID),
		fmt.Sprintf("%s/%s.json", dataDirectory, userUUID),
		// Not sure if this one is in use
		fmt.Sprintf("%s/alistsbyuser/%s.json", publishDirectory, userUUID),
		fmt.Sprintf("%s/alistsbyuser/%s.html", publishDirectory, userUUID),
	}

	h.deleteFiles(files)
	// TODO do I need to return anything?
	return nil
}
