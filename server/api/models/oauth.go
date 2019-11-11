package models

import "github.com/freshteapot/learnalist-api/server/pkg/oauth"

func (dal *DAL) OAuthHandler() oauth.OAuthReadWriter {
	return dal.oauthHandler
}
