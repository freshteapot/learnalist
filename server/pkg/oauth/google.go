package oauth

import (
	"context"
	"encoding/json"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleClient struct {
	config *oauth2.Config
}

type GoogleConfig struct {
	Key    string
	Secret string
	Server string
}

type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Link          string `json:"link"`
	Picture       string `json:"picture"`
	Gender        string `json:"gender"`
	Locale        string `json:"locale"`
	Hd            string `json:"hd"`
}

func NewGoogle(conf GoogleConfig) OAuth2ConfigInterface {
	return &googleClient{
		config: &oauth2.Config{
			RedirectURL:  conf.Server + "/api/v1/oauth/google/callback",
			ClientID:     conf.Key,
			ClientSecret: conf.Secret,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.profile",
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		},
	}
}

func (c googleClient) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return c.config.AuthCodeURL(state, opts...)
}

func (c googleClient) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return c.config.Exchange(ctx, code, opts...)
}

func (c googleClient) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return oauth2.NewClient(ctx, c.config.TokenSource(ctx, t))
}

func GoogleConvertRawUserInfo(raw []byte) (GoogleUserInfo, error) {
	var info GoogleUserInfo
	err := json.Unmarshal(raw, &info)
	return info, err
}
