package api

import (
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

func NewManager(
	datastore models.Datastore,
	userManagement user.Management,
	acl acl.Acl,
	databaseName string,
	hugoHelper hugo.HugoSiteBuilder,
	oauthHandlers oauth.Handlers,
	logger *logrus.Logger,
) *Manager {
	return &Manager{
		Datastore:      datastore,
		userManagement: userManagement,
		Acl:            acl,
		DatabaseName:   databaseName,
		HugoHelper:     hugoHelper,
		OauthHandlers:  oauthHandlers,
		logger:         logger,
		insights:       event.NewInsights(logger),
	}
}
