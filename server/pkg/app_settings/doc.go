package app_settings

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
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

func GetRemindV1(repo user.ManagementStorage, userUUID string) (openapi.AppSettingsRemindV1, error) {
	data, err := repo.GetInfo(userUUID)
	response := openapi.AppSettingsRemindV1{}

	if err != nil {
		return response, err
	}

	var pref user.UserPreference
	err = json.Unmarshal(data, &pref)
	if err != nil {
		return response, nil
	}
	if pref.Apps.RemindV1 != nil {
		response = *pref.Apps.RemindV1
	}

	return response, nil
}

func SaveRemindV1(repo user.ManagementStorage, userUUID string, settings openapi.AppSettingsRemindV1) error {
	pref := user.UserPreference{
		Apps: &user.UserPreferenceApps{
			RemindV1: &settings,
		},
	}

	b, _ := json.Marshal(pref)
	return repo.SaveInfo(userUUID, b)
}
