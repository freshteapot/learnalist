package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
)

func (dal *DAL) GetAllListsByUser(userUUID string) []alist.ShortInfo {
	return dal.alist.GetAllListsByUser(userUUID)
}

// GetListsByUserWithFilters Filter a list and return an array of lists.
// Filter by:
// - userUUID
// - userUUID, labels
// - userUUID, listType
// - userUUID, labels, listType
// Validation
// - labels needs can be single or "," separated.
// - uuid = User.Uuid
// - listType = one of the types in alist (but if its not there, it will clearly not return anything.)
func (dal *DAL) GetListsByUserWithFilters(uuid string, labels string, listType string) []alist.Alist {
	return dal.alist.GetListsByUserWithFilters(uuid, labels, listType)
}

// GetAlist Get alist
func (dal *DAL) GetAlist(uuid string) (alist.Alist, error) {
	return dal.Alist().GetAlist(uuid)
}

func (dal *DAL) RemoveAlist(alist_uuid string, user_uuid string) error {
	aList, err := dal.GetAlist(alist_uuid)

	if err != nil {
		return err
	}

	if aList.User.Uuid != user_uuid {
		return errors.New(i18n.InputDeleteAlistOperationOwnerOnly)
	}

	dal.Labels().RemoveLabelsForAlist(alist_uuid)
	err = dal.Alist().RemoveAlist(alist_uuid, user_uuid)

	if err != nil {
		log.Println(fmt.Sprintf(i18n.InternalServerErrorTalkingToDatabase, "RemoveAlist"))
		log.Println(err)
	}

	dal.Acl.DeleteList(aList.Uuid)
	return err
}

/*
If empty user, we reject.
If POST, we enforce a new uuid for the list.
If empty uuid, we reject.
If PUT, do a lookup to see if the list exists.
*/
func (dal *DAL) SaveAlist(method string, aList alist.Alist) (alist.Alist, error) {
	var err error
	// var jsonBytes []byte
	var emptyAlist alist.Alist
	err = alist.Validate(aList)
	if err != nil {
		if err == alist.ErrorSharingNotAllowedWithFrom {
			return emptyAlist, i18n.ErrorInputSaveAlistOperationFromRestriction
		}
		return emptyAlist, err
	}

	if aList.User.Uuid == "" {
		return emptyAlist, errors.New(i18n.InternalServerErrorMissingUserUuid)
	}

	// We set the uuid
	if method == http.MethodPost {
		user := &uuid.User{
			Uuid: aList.User.Uuid,
		}
		playList := uuid.NewPlaylist(user)
		aList.Uuid = playList.Uuid
	}

	if aList.Uuid == "" {
		return emptyAlist, errors.New(i18n.InternalServerErrorMissingAlistUuid)
	}

	if method == http.MethodPut {
		current, _ := dal.Alist().GetAlist(aList.Uuid)
		if current.Uuid == "" {
			return emptyAlist, errors.New(i18n.SuccessAlistNotFound)
		}

		if current.User.Uuid != aList.User.Uuid {
			return emptyAlist, errors.New(i18n.InputSaveAlistOperationOwnerOnly)
		}

		if !alist.WithFromCanUpdate(aList.Info, current.Info) {
			return emptyAlist, i18n.ErrorInputSaveAlistOperationFromModify
		}

		// Check if what is about to be written is the same.
		a, _ := json.Marshal(&aList)
		b, _ := json.Marshal(current)
		if string(a) == string(b) {
			return aList, nil
		}
	}

	dal.Labels().RemoveLabelsForAlist(aList.Uuid)
	err = dal.SaveLabelsForAlist(aList)
	if err != nil {
		log.Println(err)
	}

	// This is the only part that would need handling
	_, err = dal.Alist().SaveAlist(method, aList)

	if err != nil {
		return emptyAlist, err
	}

	// Set shared
	switch aList.Info.SharedWith {
	case aclKeys.SharedWithPublic:
		dal.Acl.ShareListWithPublic(aList.Uuid)
	case aclKeys.SharedWithFriends:
		dal.Acl.ShareListWithFriends(aList.Uuid)
	case aclKeys.NotShared:
		fallthrough
	default:
		dal.Acl.MakeListPrivate(aList.Uuid, aList.User.Uuid)
	}

	return aList, nil
}

// Process the lists labels,
// We post to both the user_labels and alist_labels table.
func (dal *DAL) SaveLabelsForAlist(aList alist.Alist) error {
	// Post the labels
	var statusCode int
	var err error

	for _, input := range aList.Info.Labels {
		a := label.NewUserLabel(input, aList.User.Uuid)
		b := label.NewAlistLabel(input, aList.User.Uuid, aList.Uuid)

		statusCode, err = dal.Labels().PostUserLabel(a)
		if statusCode == http.StatusBadRequest {
			return err
		}

		statusCode, err = dal.Labels().PostAlistLabel(b)
		if statusCode == http.StatusBadRequest {
			return err
		}
	}
	return err
}
