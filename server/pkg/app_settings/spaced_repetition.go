package app_settings

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func GetSpacedRepetition(repo user.ManagementStorage, userUUID string) (user.SpacedRepetition, error) {
	data, err := repo.GetInfo(userUUID)

	info := user.SpacedRepetition{}

	if err != nil {
		// Can return utils.ErrNotFound, based on repo.GetInfo
		return info, err
	}

	var pref user.UserPreference
	err = json.Unmarshal(data, &pref)
	if err != nil {
		return info, nil
	}

	// This was important
	if pref.SpacedRepetition == nil {
		return info, utils.ErrNotFound
	}

	info = *pref.SpacedRepetition
	return info, nil
}

func SaveSpacedRepetition(repo user.ManagementStorage, userUUID string, spacedRepetition user.SpacedRepetition) error {
	/*
		if len(spacedRepetition.ListsOvertime) == 0 {
			key := "spaced_repetition"
			return repo.RemoveInfo(userUUID, key)
		}
	*/
	pref := user.UserPreference{
		SpacedRepetition: &spacedRepetition,
	}

	b, _ := json.Marshal(pref)
	return repo.SaveInfo(userUUID, b)
}
