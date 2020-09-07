package models

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/jmoiron/sqlx"
)

// DB allowing us to build an abstraction layer
type DAL struct {
	Db                          *sqlx.DB
	Acl                         acl.Acl
	userSession                 user.Session
	userFromIDP                 user.UserFromIDP
	userWithUsernameAndPassword user.UserWithUsernameAndPassword
	oauthHandler                oauth.OAuthReadWriter
	labels                      label.LabelReadWriter
	alist                       alist.DatastoreAlists
}

func NewDAL(db *sqlx.DB, acl acl.Acl, aListStorage alist.DatastoreAlists, labels label.LabelReadWriter, userSession user.Session, userFromIDP user.UserFromIDP, userWithUsernameAndPassword user.UserWithUsernameAndPassword, oauthHandler oauth.OAuthReadWriter) *DAL {

	dal := &DAL{
		Db:                          db,
		Acl:                         acl,
		userSession:                 userSession,
		userFromIDP:                 userFromIDP,
		userWithUsernameAndPassword: userWithUsernameAndPassword,
		oauthHandler:                oauthHandler,
		labels:                      labels,
		alist:                       aListStorage,
	}
	return dal
}

func (dal *DAL) Alist() DatastoreAlists {
	return dal.alist
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
		aList, err = dal.Alist().GetAlist(uuid)
		found := utils.StringArrayIndexOf(aList.Info.Labels, label)
		if found != -1 {
			cleaned := []string{}
			for _, item := range aList.Info.Labels {
				if item != label {
					cleaned = append(cleaned, item)
				}
			}
			aList.Info.Labels = cleaned
			dal.Alist().SaveAlist(http.MethodPut, aList)
		}
	}

	return dal.Labels().RemoveUserLabel(label, user)
}

func (dal *DAL) GetPublicLists() []alist.ShortInfo {
	return dal.Alist().GetPublicLists()
}
