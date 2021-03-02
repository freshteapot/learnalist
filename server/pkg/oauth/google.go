package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/tideland/gorest/jwt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

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
		audiences: conf.Audiences,
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

func (c googleClient) GetUserUUIDFromIDP(input IDPOauthInput) (string, error) {
	j, err := jwt.Decode(input.IDToken)
	if err != nil {
		return "", errors.New("bad token")
	}

	leeway := time.Minute
	if !j.IsValid(leeway) {
		return "", errors.New("time has passed")
	}

	iss, _ := j.Claims().GetString("iss")

	if iss != "https://accounts.google.com" {
		return "", errors.New("bad-issuer")
	}

	aud, _ := j.Claims().GetString("aud")
	sub, _ := j.Claims().GetString("sub")

	// TODO check audience
	match := false
	for _, supported := range c.audiences {
		if supported == aud {
			match = true
			break
		}
	}

	if !match {
		return "", fmt.Errorf("%s audience not on the list for google", aud)
	}
	// At this point, life could go on
	return sub, nil
}

func GoogleConvertRawUserInfo(raw []byte) (GoogleUserInfo, error) {
	var info GoogleUserInfo
	err := json.Unmarshal(raw, &info)
	return info, err
}
