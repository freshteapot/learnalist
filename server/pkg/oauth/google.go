package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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

func (c googleClient) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	return oauth2.NewClient(ctx, c.config.TokenSource(ctx, t))
}

func (c googleClient) GetUserUUIDFromIDP(input IDPOauthInput) (string, error) {
	//// I dont need to use this, if I am happy to consume the jwt token
	//// TODO set defaults
	//httpClient := &http.Client{
	//	Timeout: 5 * time.Second,
	//}
	//
	//oauth2Service, err := googleOauth2Api.New(httpClient)
	//tokenInfoCall := oauth2Service.Tokeninfo()
	//tokenInfoCall.IdToken(input.IDToken)
	//
	//tokenInfo, err := tokenInfoCall.Do()
	//if err != nil {
	//	return "", err
	//}
	//
	//// Check the audience
	//if !utils.StringArrayContains(c.audiences, tokenInfo.Audience) {
	//	return "", fmt.Errorf("%s audience not on the list for google", tokenInfo.Audience)
	//}
	//
	//return tokenInfo.UserId, nil
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

// TODO idp specific
// TODO needs better http setup
//req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo?prettyprint=false", nil)
//req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", input.AccessToken))
//resp, err := http.DefaultClient.Do(req)
//if err != nil {
//	logContext.WithFields(logrus.Fields{
//		"event": "idp-user-info-via-google-1",
//		"error": err,
//	}).Error("Issue in login via idp")
//	return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
//}
//
//defer resp.Body.Close()
//if resp.StatusCode != http.StatusOK {
//	logContext.WithFields(logrus.Fields{
//		"event":       "idp-user-info-via-google-2",
//		"status_code": resp.StatusCode,
//		"error":       err,
//	}).Error("Issue in login via idp")
//	return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
//}
//
//contents, err := ioutil.ReadAll(resp.Body)
//if err != nil {
//	logContext.WithFields(logrus.Fields{
//		"event": "idp-user-info-via-google-3",
//		"error": err,
//	}).Error("Issue in login via idp")
//	return c.JSON(http.StatusForbidden, api.HTTPAccessDeniedResponse)
//}
