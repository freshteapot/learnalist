package app_settings

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func GetRemindV1(repo user.ManagementStorage, userUUID string) (openapi.AppSettingsRemindV1, error) {
	data, err := repo.GetInfo(userUUID)
	settings := openapi.AppSettingsRemindV1{}

	if err != nil {
		if err != utils.ErrNotFound {
			return settings, err
		}

		// Assume not set, and default to true
		settings.SpacedRepetition.PushEnabled = 1
		return settings, nil
	}

	var pref user.UserPreference
	err = json.Unmarshal(data, &pref)
	if err != nil {
		return settings, nil
	}

	if pref.Apps.RemindV1 != nil {
		settings = *pref.Apps.RemindV1
	}

	return settings, nil
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
