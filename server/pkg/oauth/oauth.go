package oauth

import (
	"golang.org/x/oauth2"
)

type Handlers struct {
	Google *oauth2.Config
}

type OAuthReadWriter interface {
	GetTokenInfo(userUUID string) (*oauth2.Token, error)
	WriteTokenInfo(userUUID string, token *oauth2.Token) error
}
