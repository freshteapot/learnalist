package info

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func getSpacedRepetition(repo user.ManagementStorage, userUUID string) (user.SpacedRepetition, error) {
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

func saveSpacedRepetition(repo user.ManagementStorage, userUUID string, spacedRepetition user.SpacedRepetition) error {
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

func AppendAndSaveSpacedRepetition(repo user.ManagementStorage, userUUID string, alistUUID string) error {
	info, err := getSpacedRepetition(repo, userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			return err
		}
	}

	if utils.StringArrayContains(info.ListsOvertime, alistUUID) {
		return nil
	}

	info.ListsOvertime = append(info.ListsOvertime, alistUUID)
	return saveSpacedRepetition(repo, userUUID, info)
}

func RemoveAndSaveSpacedRepetition(repo user.ManagementStorage, userUUID string, alistUUID string) error {
	info, err := getSpacedRepetition(repo, userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			return err
		}
		return nil
	}

	found := utils.StringArrayIndexOf(info.ListsOvertime, alistUUID)

	if found == -1 {
		return nil
	}

	info.ListsOvertime = utils.StringArrayRemoveAtIndex(info.ListsOvertime, found)
	return saveSpacedRepetition(repo, userUUID, info)
}
