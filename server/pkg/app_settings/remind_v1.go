package app_settings

import (
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func GetRemindV1(repo user.UserInfoRepository, userUUID string) (openapi.AppSettingsRemindV1, error) {
	pref, err := repo.Get(userUUID)

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

	// This was important
	if pref.Apps == nil {
		return settings, utils.ErrNotFound
	}

	if pref.Apps.RemindV1 != nil {
		settings = *pref.Apps.RemindV1
	}

	return settings, nil
}

func SaveRemindV1(repo user.UserInfoRepository, userUUID string, settings openapi.AppSettingsRemindV1) error {
	pref, err := repo.Get(userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			return err
		}
	}

	if pref.Apps == nil {
		pref.Apps = &user.UserPreferenceApps{}
	}

	pref.Apps.RemindV1 = &settings
	return repo.Save(userUUID, pref)
}
