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

type HttpLabelInput struct {
	Label string `json:"label"`
}

type HttpGetVersionResponse struct {
	GitHash string `json:"gitHash"`
	GitDate string `json:"gitDate"`
	Version string `json:"version"`
	Url     string `json:"url"`
}

type HttpShareListInput struct {
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

type HttpShareListWithUserInput struct {
	UserUUID  string `json:"user_uuid"`
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

// m exposing the data abstraction layer

type HttpResponseMessage struct {
	Message string `json:"message"`
}

type Manager struct {
	Datastore      models.Datastore
	userManagement user.Management
	Acl            acl.Acl
	DatabaseName   string
	HugoHelper     hugo.HugoSiteBuilder
	OauthHandlers  oauth.Handlers
	logger         *logrus.Logger
	insights       event.Insights
}
