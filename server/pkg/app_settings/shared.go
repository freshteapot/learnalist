package app_settings

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/user"
)

func GetAll(repo user.ManagementStorage, userUUID string) (interface{}, error) {
	data, err := repo.GetInfo(userUUID)
	var response interface{}

	if err != nil {
		return response, err
	}

	var pref user.UserPreference
	err = json.Unmarshal(data, &pref)
	if err != nil {
		return response, nil
	}

	return pref.Apps, nil
}
