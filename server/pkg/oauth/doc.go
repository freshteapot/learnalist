package oauth

import (
	"context"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"golang.org/x/oauth2"
)

const (
	IDPKeyGoogle = "google"
	IDPKeyApple  = "apple"
)

type Handlers struct {
	keys    []string
	Google  OAuth2ConfigInterface
	AppleID OAuth2ConfigInterface
}

type IDPTokeninfo struct {
	// Audience: Who is the intended audience for this token. In general the
	// same as issued_to.
	Aud string `json:"aud,omitempty"`
	// UserId: The obfuscated user id.
	Sub string `json:"sub,omitempty"`
}

// https://blog.seriesci.com/how-to-mock-oauth-in-go/
type OAuth2ConfigInterface interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
	// Return the extID / userID from the idp
	GetUserUUIDFromIDP(input openapi.HttpUserLoginIdpInput) (string, error)
}

type googleClient struct {
	config    *oauth2.Config
	audiences []string
}

type GoogleConfig struct {
	Key       string
	Secret    string
	Server    string
	Audiences []string
}

type appleClient struct {
	redirectURI string
	webAudience AppleConfig
	audiences   []AppleConfig
}

type AppleConfig struct {
	TeamID   string
	ClientID string
	KeyID    string
	Server   string
	Cert     string
}
