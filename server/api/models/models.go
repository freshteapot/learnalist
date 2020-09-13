package models

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/label"
	apiUser "github.com/freshteapot/learnalist-api/server/api/user"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Acl                         acl.Acl
	userSession                 user.Session
	userFromIDP                 user.UserFromIDP
	userWithUsernameAndPassword user.UserWithUsernameAndPassword
	oauthHandler                oauth.OAuthReadWriter
	labels                      label.LabelReadWriter
	alist                       alist.DatastoreAlists
	user                        apiUser.DatastoreUsers
}

func NewDAL(acl acl.Acl, apiUserStorage apiUser.DatastoreUsers, aListStorage alist.DatastoreAlists, labels label.LabelReadWriter, userSession user.Session, userFromIDP user.UserFromIDP, userWithUsernameAndPassword user.UserWithUsernameAndPassword, oauthHandler oauth.OAuthReadWriter) *DAL {
	dal := &DAL{
		Acl:                         acl,
		userSession:                 userSession,
		userFromIDP:                 userFromIDP,
		userWithUsernameAndPassword: userWithUsernameAndPassword,
		oauthHandler:                oauthHandler,
		labels:                      labels,
		alist:                       aListStorage,
		user:                        apiUserStorage,
	}
	return dal
}

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
		aList, err = dal.alist.GetAlist(uuid)
		if err != nil {
			if err == i18n.ErrorListNotFound {
				continue
			}
			// TODO this is not ideal
			panic(err)
		}

		found := utils.StringArrayIndexOf(aList.Info.Labels, label)
		if found != -1 {
			cleaned := []string{}
			for _, item := range aList.Info.Labels {
				if item != label {
					cleaned = append(cleaned, item)
				}
			}
			aList.Info.Labels = cleaned
			dal.alist.SaveAlist(http.MethodPut, aList)
		}
	}

	return dal.Labels().RemoveUserLabel(label, user)
}

func (dal *DAL) GetPublicLists() []alist.ShortInfo {
	return dal.alist.GetPublicLists()
}
