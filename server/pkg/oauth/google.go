package oauth

import (
	"encoding/json"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

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

func GoogleConvertRawUserInfo(raw []byte) (GoogleUserInfo, error) {
	var info GoogleUserInfo
	err := json.Unmarshal(raw, &info)
	return info, err
}
