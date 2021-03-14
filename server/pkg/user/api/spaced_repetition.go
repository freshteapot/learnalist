package api

import (
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
)

func AppendAndSaveSpacedRepetition(repo user.UserInfoRepository, userUUID string, alistUUID string) error {
	pref, err := repo.Get(userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			return err
		}
	}

	if pref.SpacedRepetition == nil {
		pref.SpacedRepetition = &user.SpacedRepetition{}
	}

	if utils.StringArrayContains(pref.SpacedRepetition.ListsOvertime, alistUUID) {
		return nil
	}

	pref.SpacedRepetition.ListsOvertime = append(pref.SpacedRepetition.ListsOvertime, alistUUID)
	return repo.Save(userUUID, pref)
}

func RemoveAndSaveSpacedRepetition(repo user.UserInfoRepository, userUUID string, alistUUID string) error {
	pref, err := repo.Get(userUUID)
	if err != nil {
		if err != utils.ErrNotFound {
			return err
		}
		return nil
	}

	if pref.SpacedRepetition == nil {
		return nil
	}

	found := utils.StringArrayIndexOf(pref.SpacedRepetition.ListsOvertime, alistUUID)

	if found == -1 {
		return nil
	}

	pref.SpacedRepetition.ListsOvertime = utils.StringArrayRemoveAtIndex(pref.SpacedRepetition.ListsOvertime, found)
	return repo.Save(userUUID, pref)
}
