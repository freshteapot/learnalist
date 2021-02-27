package oauth

import (
	"context"
	"net/http"

	"golang.org/x/oauth2"
)

type Handlers struct {
	Google  OAuth2ConfigInterface
	AppleID OAuth2ConfigInterface
}

// https://blog.seriesci.com/how-to-mock-oauth-in-go/
type OAuth2ConfigInterface interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	Client(ctx context.Context, t *oauth2.Token) *http.Client
}

type OAuthReadWriter interface {
	GetTokenInfo(userUUID string) (*oauth2.Token, error)
	WriteTokenInfo(userUUID string, token *oauth2.Token) error
}
