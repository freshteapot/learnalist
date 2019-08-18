package hugo

import "fmt"

type HugoHelper struct {
	Cwd              string
	DataDirectory    string
	ContentDirectory string
}

func NewHugoHelper(cwd string) *HugoHelper {
	// TODO make sure the dataDir exists
	dataDirectory := fmt.Sprintf("%s/data/lists", cwd)
	contentDirectory := fmt.Sprintf("%s/content/alists", cwd)

	return &HugoHelper{
		Cwd:              cwd,
		DataDirectory:    dataDirectory,
		ContentDirectory: contentDirectory,
	}
}
