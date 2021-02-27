package oauth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/Timothylock/go-signin-with-apple/apple"
	"github.com/tideland/gorest/jwt"
	"golang.org/x/oauth2"
)

type appleClient struct {
	config      AppleConfig
	redirectURI string
}

type AppleConfig struct {
	TeamID   string
	ClientID string
	KeyID    string
	Server   string
	Cert     string
}

func NewAppleID(conf AppleConfig) OAuth2ConfigInterface {
	return &appleClient{
		config:      conf,
		redirectURI: conf.Server + "/api/v1/oauth/appleid/callback",
	}
}

func (c *appleClient) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	u := url.Values{}
	u.Add("response_type", "code")
	u.Add("redirect_uri", c.redirectURI)
	u.Add("client_id", c.config.ClientID)
	u.Add("state", state)
	return "https://appleid.apple.com/auth/authorize?" + u.Encode()
}

func (c *appleClient) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	secret, _ := apple.GenerateClientSecret(c.config.Cert, c.config.TeamID, c.config.ClientID, c.config.KeyID)
	// Generate a new validation client
	client := apple.New()

	vReq := apple.AppValidationTokenRequest{
		ClientID:     c.config.ClientID,
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

	if aud != c.config.ClientID {
		return nil, errors.New("bad-client-not-match")
	}

	// TODO create date from ExpiresIN
	t := &oauth2.Token{
		AccessToken:  accessToken,
		TokenType:    tokenType,
		RefreshToken: refreshToken,
		Expiry:       time.Now(),
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
