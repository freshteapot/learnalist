package oauth

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleConfig struct {
	Key    string
	Secret string
	Server string
}

func NewGoogle(conf GoogleConfig) *oauth2.Config {
	googleConfig := &oauth2.Config{
		RedirectURL:  conf.Server + "/api/v1/oauth/google/callback",
		ClientID:     conf.Key,
		ClientSecret: conf.Secret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}
	return googleConfig
}
