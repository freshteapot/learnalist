package authenticate

import (
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

const (
	defaultRealm = "Restricted"
)

var LookupBasic func(username string, hash string) (string, error)
var LookupBearer func(token string) (string, error)
var SkipAuth func(c echo.Context) bool

func Auth(next echo.HandlerFunc) echo.HandlerFunc {

	if SkipAuth == nil {
		panic("You need to set SkipAuth")
	}

	if LookupBearer == nil {
		panic("You need to set LookupBearer")
	}

	if LookupBearer == nil {
		panic("You need to set LookupBasic")
	}

	return func(c echo.Context) error {
		var valid bool
		var err error

		if SkipAuth(c) {
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

func validateBasic(c echo.Context, basic string) (bool, error) {
	b, err := base64.StdEncoding.DecodeString(basic)
	if err != nil {
		return false, nil
	}

	parts := strings.SplitN(string(b), ":", 2)
	if len(parts) != 2 {
		return false, nil
	}

	username := parts[0]
	password := parts[1]
	hash := HashIt(username, password)

	// TODO this is ugly
	userUUID, err := LookupBasic(username, hash)
	if err != nil {
		return false, nil
	}

	user := &uuid.User{
		Uuid: userUUID,
	}
	c.Set("loggedInUser", *user)
	return true, nil
}

// TODO lets use real encryption ;)
func HashIt(username string, password string) string {
	h := fnv.New64()
	beforeHash := fmt.Sprintf("%s:%s", username, password)

	h.Write([]byte(beforeHash))
	hash := h.Sum64()
	storedHash := fmt.Sprintf("%d", hash)
	return storedHash
}
