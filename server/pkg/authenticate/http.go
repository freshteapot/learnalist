package authenticate

import (
	"fmt"
	"net/http"
	"time"
)

type CookieConfig struct {
	Domain string
	Secure bool
}

var config CookieConfig

func SetLoginCookieConfig(_config CookieConfig) {
	config = _config
}

func NewLoginCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:     "x-authentication-bearer",
		Value:    token,
		Path:     "/",
		Domain:   fmt.Sprintf(".%s", config.Domain),
		HttpOnly: true,
		Secure:   config.Secure,
		Expires:  time.Now().UTC().Add(30 * (time.Hour * 24)),
	}
}
