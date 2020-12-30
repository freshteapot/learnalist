package app_settings

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func GetRemindV1(repo user.ManagementStorage, userUUID string) (openapi.AppSettingsRemindV1, error) {
	data, err := repo.GetInfo(userUUID)

	// Assume not set, and default to true
	settings := openapi.AppSettingsRemindV1{
		SpacedRepetition: openapi.AppSettingsRemindV1SpacedRepetition{
			PushEnabled: 1,
		},
	}

	if err != nil {
		// Can return utils.ErrNotFound, based on repo.GetInfo
		return settings, err
	}

	var pref user.UserPreference
	err = json.Unmarshal(data, &pref)
	if err != nil {
		return settings, nil
	}

	// This was important
	if pref.Apps == nil {
		return settings, utils.ErrNotFound
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
