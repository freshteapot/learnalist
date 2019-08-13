package alist

import "encoding/json"

func parseAlistInfo(jsonBytes []byte) (AlistInfo, error) {
	listInfo := new(AlistInfo)
	err := json.Unmarshal(jsonBytes, &listInfo)
	if listInfo.Labels == nil {
		listInfo.Labels = []string{}
	}
	return *listInfo, err
}
