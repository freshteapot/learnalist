package api

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

type OauthService struct {
	userManagement user.Management
	hugoHelper     hugo.HugoSiteBuilder
	oauthHandlers  oauth.Handlers
	logContext     logrus.FieldLogger
	userSession    user.Session
	userFromIDP    user.UserFromIDP
	alistsRepo     alist.DatastoreAlists
}

func NewService(
	userManagement user.Management,
	hugoHelper hugo.HugoSiteBuilder,
	oauthHandlers oauth.Handlers,
	userSession user.Session,
	userFromIDP user.UserFromIDP,
	alistsRepo alist.DatastoreAlists,
	logContext logrus.FieldLogger,
) OauthService {
	return OauthService{
		userManagement: userManagement,
		userSession:    userSession,
		userFromIDP:    userFromIDP,
		hugoHelper:     hugoHelper,
		alistsRepo:     alistsRepo,
		oauthHandlers:  oauthHandlers,
		logContext:     logContext,
	}
}
