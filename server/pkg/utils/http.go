package utils

import (
	"errors"
	"net/http"
)

func GetCookieByName(cookie []*http.Cookie, name string) (*http.Cookie, error) {
	cookieLen := len(cookie)
	for i := 0; i < cookieLen; i++ {
		if cookie[i].Name == name {
			return cookie[i], nil
		}
	}
	return nil, errors.New("not-found")
}
