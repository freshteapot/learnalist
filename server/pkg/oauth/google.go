package oauth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/tideland/gorest/jwt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

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

func (c googleClient) GetUserUUIDFromIDP(input openapi.HttpUserLoginIdpInput) (string, error) {
	j, err := jwt.Decode(input.IdToken)
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
	return sub, nil
}
