package oauth

import (
	"context"
	"net/http"

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

type IDPOauthInput struct {
	Idp     string `json:"idp"`
	IDToken string `json:"id_token"`
	Code    string `json:"code,omitempty"`
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
	Client(ctx context.Context, t *oauth2.Token) *http.Client
	// Return the extID / userID from the idp
	GetUserUUIDFromIDP(input IDPOauthInput) (string, error)
}

type OAuthReadWriter interface {
	GetTokenInfo(userUUID string) (*oauth2.Token, error)
	WriteTokenInfo(userUUID string, token *oauth2.Token) error
}

func NewHandlers() *Handlers {
	return &Handlers{}
}

func (h *Handlers) AddGoogle(handler OAuth2ConfigInterface) {
	h.keys = append(h.keys, IDPKeyGoogle)
	h.Google = handler
}

func (h *Handlers) AddAppleID(handler OAuth2ConfigInterface) {
	h.keys = append(h.keys, IDPKeyApple)
	h.AppleID = handler
}

func (h *Handlers) Keys() []string {
	return h.keys
}
