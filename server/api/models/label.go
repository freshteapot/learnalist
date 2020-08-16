package models

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/api/utils"
)

func (dal *DAL) Labels() label.LabelReadWriter {
	return dal.labels
}

// Pass in the label and the user (uuid) to remove them from the tables
func (dal *DAL) RemoveUserLabel(label string, user string) error {

	var (
		err   error
		aList alist.Alist
		uuids []string
	)
	uuids, err = dal.Labels().GetUniqueListsByUserAndLabel(label, user)
	if err != nil {
		return err
	}

	for _, uuid := range uuids {
		aList, err = dal.GetAlist(uuid)
		found := utils.StringArrayIndexOf(aList.Info.Labels, label)
		if found != -1 {
			cleaned := []string{}
			for _, item := range aList.Info.Labels {
				if item != label {
					cleaned = append(cleaned, item)
				}
			}
			aList.Info.Labels = cleaned
			dal.SaveAlist(http.MethodPut, aList)
		}
	}

	return dal.Labels().RemoveUserLabel(label, user)
}
