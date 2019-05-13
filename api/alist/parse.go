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
func parseAlistTypeV1(jsonBytes []byte) (AlistTypeV1, error) {
	listData := new(AlistTypeV1)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}

func parseAlistTypeV2(jsonBytes []byte) (AlistTypeV2, error) {
	listData := new(AlistTypeV2)
	err := json.Unmarshal(jsonBytes, &listData)
	return *listData, err
}
