package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Timothylock/go-signin-with-apple/apple"
	"github.com/tideland/gorest/jwt"
	"golang.org/x/oauth2"
)

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

func NewAppleID(webAudience AppleConfig, audiences []AppleConfig) OAuth2ConfigInterface {
	return &appleClient{
		webAudience: webAudience,
		redirectURI: webAudience.Server + "/api/v1/oauth/appleid/callback",
		audiences:   audiences,
	}
}

func (c *appleClient) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	u := url.Values{}
	u.Add("response_type", "code")
	u.Add("redirect_uri", c.redirectURI)
	u.Add("client_id", c.webAudience.ClientID)
	u.Add("state", state)
	return "https://appleid.apple.com/auth/authorize?" + u.Encode()
}

func (c *appleClient) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	// This one needs to use the web
	config := c.webAudience
	secret, _ := apple.GenerateClientSecret(config.Cert, config.TeamID, config.ClientID, config.KeyID)
	// Generate a new validation client
	client := apple.New()

	vReq := apple.AppValidationTokenRequest{
		ClientID:     config.ClientID,
		ClientSecret: secret,
		Code:         code,
	}

	var resp apple.ValidationResponse

	// Do the verification
	err := client.VerifyAppToken(context.Background(), vReq, &resp)
	if err != nil {
		return nil, err
	}

	idToken := resp.IDToken
	accessToken := resp.AccessToken
	tokenType := resp.TokenType
	refreshToken := resp.RefreshToken

	j, err := jwt.Decode(idToken)
	if err != nil {
		return nil, errors.New("bad-jwt")
	}

	iss, _ := j.Claims().GetString("iss")
	aud, _ := j.Claims().GetString("aud")
	sub, _ := j.Claims().GetString("sub")

	// I wonder if I need to test this here as I am talking to apple
	if iss != "https://appleid.apple.com" {
		return nil, errors.New("bad-issuer")
	}

	if aud != config.ClientID {
		return nil, errors.New("bad-client-not-match")
	}

	// TODO create date from ExpiresIN
	t := &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    tokenType,
		RefreshToken: refreshToken,
		Expiry:       time.Now(), // TODO change
	}

	t = t.WithExtra(map[string]interface{}{
		"id_token": idToken,
		"aud":      aud,
		"sub":      sub,
		"iss":      iss,
	})
	return t, nil
}

func (c *appleClient) Client(ctx context.Context, t *oauth2.Token) *http.Client {
	// TODO this is not in use
	return &http.Client{}
}

func (c *appleClient) GetUserUUIDFromIDP(input IDPOauthInput) (string, error) {

	// Or we go with "github.com/Timothylock/go-signin-with-apple/apple"
	j, err := jwt.Decode(input.IDToken)
	if err != nil {
		return "", errors.New("bad token")
	}

	leeway := time.Minute
	if !j.IsValid(leeway) {
		return "", errors.New("time has passed")
	}

	iss, _ := j.Claims().GetString("iss")

	if iss != "https://appleid.apple.com" {
		return "", errors.New("bad-issuer")
	}

	aud, _ := j.Claims().GetString("aud")
	sub, _ := j.Claims().GetString("sub")

	// TODO check audience
	match := false
	var config AppleConfig
	for _, supported := range c.audiences {
		if supported.ClientID == aud {
			match = true
			config = supported
			break
		}
	}

	if !match {
		return "", fmt.Errorf("%s audience not on the list for appleid", aud)
	}
	// At this point, life could go on
	//return sub, nil
	// Setting the below up will mean replay of id_token wont work, it does with google
	// TODO do I want to further lock it down?
	// I wonder if these actually need callbacks setup?
	secret, _ := apple.GenerateClientSecret(config.Cert, config.TeamID, config.ClientID, config.KeyID)
	client := apple.New()

	vReq := apple.AppValidationTokenRequest{
		ClientID:     config.ClientID,
		ClientSecret: secret,
		Code:         input.Code,
	}

	var resp apple.ValidationResponse

	// Do the verification
	err = client.VerifyAppToken(context.Background(), vReq, &resp)
	if err != nil {
		return "", err
	}

	return sub, nil
}
