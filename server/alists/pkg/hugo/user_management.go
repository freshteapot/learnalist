package hugo

import (
	"fmt"
)

func (h HugoHelper) DeleteList(listUUID string) error {
	contentDirectory := fmt.Sprintf(RealtivePathContentAlist, h.cwd)
	dataDirectory := fmt.Sprintf(RealtivePathDataAlist, h.cwd)
	publishDirectory := fmt.Sprintf(RealtivePathPublicContentAlist, h.cwd)

	files := []string{
		fmt.Sprintf("%s/%s.md", contentDirectory, listUUID),
		fmt.Sprintf("%s/%s.json", dataDirectory, listUUID),
		// Not sure if this one is in use
		fmt.Sprintf("%s/%s.json", publishDirectory, listUUID),
		fmt.Sprintf("%s/%s.html", publishDirectory, listUUID),
	}

	h.deleteFiles(files)
	return nil
}

func (h HugoHelper) DeleteUser(userUUID string) error {
	contentDirectory := fmt.Sprintf(RealtivePathContentAlistsByUser, h.cwd)
	dataDirectory := fmt.Sprintf(RealtivePathDataAlistsByUser, h.cwd)
	publishDirectory := fmt.Sprintf(RealtivePathPublicContentAlistsByUser, h.cwd)

	files := []string{
		fmt.Sprintf("%s/%s.md", contentDirectory, userUUID),
		fmt.Sprintf("%s/%s.json", dataDirectory, userUUID),
		// Not sure if this one is in use
		fmt.Sprintf("%s/%s.json", publishDirectory, userUUID),
		fmt.Sprintf("%s/%s.html", publishDirectory, userUUID),
	}

	h.deleteFiles(files)
	return nil
}
