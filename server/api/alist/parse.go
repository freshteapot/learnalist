package alist

import (
	"encoding/json"

	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
)

func parseAlistInfo(jsonBytes []byte) (AlistInfo, error) {
	listInfo := new(AlistInfo)
	err := json.Unmarshal(jsonBytes, &listInfo)
	if listInfo.Labels == nil {
		listInfo.Labels = []string{}
	}
	if listInfo.SharedWith == "" {
		listInfo.SharedWith = aclKeys.NotShared
	}
	return *listInfo, err
}
