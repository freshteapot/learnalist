package authenticate

import (
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

const (
	defaultRealm = "Restricted"
)

var LookupBasic func(loginUser LoginUser) (*uuid.User, error)
var LookupBearer func(token string) (string, error)

func Auth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var valid bool
		var err error

		if skip(c) {
			return next(c)
		}

		authorization := c.Request().Header.Get("Authorization")
		parts := strings.SplitN(authorization, " ", 2)
		if len(parts) != 2 {
			return echo.ErrForbidden
		}

		what := strings.ToLower(parts[0])

		switch what {
		case "basic":
			hash := parts[1]
			valid, err = validateBasic(c, hash)
		case "bearer":
			token := parts[1]
			valid, err = validateBearer(c, token)
		default:
			return echo.ErrForbidden
		}

		if err != nil {
			return echo.ErrInternalServerError
		}

		if valid {
			return next(c)
		}

		realm := defaultRealm
		// Need to return `401` for browsers to pop-up login box.
		c.Response().Header().Set(echo.HeaderWWWAuthenticate, what+" realm="+realm)
		return echo.ErrUnauthorized
	}

}

func skip(c echo.Context) bool {
	url := c.Request().URL.Path
	method := c.Request().Method
	url = strings.TrimPrefix(url, "/api/v1")

	switch method {
	case http.MethodGet:
		if url == "/" {
			return true
		}

		if url == "/version" {
			return true
		}

		if strings.HasPrefix(url, "/oauth/") {
			return true
		}
	case http.MethodPost:
		// Unfiltered ability to register a user
		if url == "/register" {
			return true
		}
	default:
		return false
	}

	return false
}

func validateBearer(c echo.Context, token string) (bool, error) {
	userUUID, err := LookupBearer(token)
	if err != nil {
		return false, nil
	}

	user := &uuid.User{
		Uuid: userUUID,
	}
	c.Set("loggedInUser", *user)
	return true, nil
}

func validateBasic(c echo.Context, hash string) (bool, error) {
	b, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false, nil
	}

	parts := strings.SplitN(string(b), ":", 2)
	if len(parts) != 2 {
		return false, nil
	}

	username := parts[0]
	password := parts[1]

	loginUser := &LoginUser{
		Username: username,
		Password: password,
	}

	user, err := LookupBasic(*loginUser)
	if err != nil {
		return false, nil
	}

	c.Set("loggedInUser", *user)
	return true, nil
}

type LoginUser struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func HashIt(user LoginUser) (string, error) {
	h := fnv.New64()
	beforeHash := fmt.Sprintf("%s:%s", user.Username, user.Password)

	h.Write([]byte(beforeHash))
	hash := h.Sum64()
	storedHash := fmt.Sprintf("%d", hash)
	return storedHash, nil
}
