package authenticate

import (
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"strings"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
)

const (
	defaultRealm = "Restricted"
)

type Config struct {
	Skip         func(c echo.Context) bool
	LookupBasic  func(username string, hash string) (string, error)
	LookupBearer func(token string) (string, error)
	// TODO maybe we should actually lookup via cookie
}

func Auth(config Config) echo.MiddlewareFunc {
	if config.LookupBearer == nil {
		panic("You need to set LookupBearer")
	}

	if config.LookupBearer == nil {
		panic("You need to set LookupBasic")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var valid bool
			var err error

			if config.Skip(c) {
				return next(c)
			}

			authorization := c.Request().Header.Get("Authorization")
			parts := strings.SplitN(authorization, " ", 2)
			if len(parts) != 2 {
				// TODO check cookie
				cookie, err := utils.GetCookieByName(c.Request().Cookies(), "x-authentication-bearer")
				if err != nil {
					return echo.ErrForbidden
				}

				// TODO how to handle non json response
				// TODO maybe use lookupCookie
				parts = []string{
					"bearer",
					cookie.Value,
				}
			}

			what := strings.ToLower(parts[0])
			switch what {
			case "basic":
				hash := parts[1]
				valid, err = config.validateBasic(c, hash)
			case "bearer":
				token := parts[1]
				valid, err = config.validateBearer(c, token)
			default:
				return echo.ErrForbidden
			}

			// Hmm currently this will never be triggered :(
			if err != nil {
				return echo.ErrInternalServerError
			}

			if valid {
				return next(c)
			}

			if what == "basic" {
				realm := defaultRealm
				// Need to return `401` for browsers to pop-up login box.
				c.Response().Header().Set(echo.HeaderWWWAuthenticate, what+" realm="+realm)
				return echo.ErrUnauthorized
			}
			return echo.ErrForbidden
		}
	}
}

func SendLogoutCookie(c echo.Context) {
	cookie := NewLoginCookie("")
	cookie.Expires = time.Now().Add(-100 * time.Hour)
	cookie.MaxAge = -1
	cookie.Value = ""

	c.SetCookie(cookie)
}

func (config Config) validateBearer(c echo.Context, token string) (bool, error) {
	userUUID, err := config.LookupBearer(token)
	if err != nil {
		SendLogoutCookie(c)
		return false, nil
	}

	user := &uuid.User{
		Uuid: userUUID,
	}
	c.Set("loggedInUser", *user)
	return true, nil
}

func (config Config) validateBasic(c echo.Context, basic string) (bool, error) {
	b, err := base64.StdEncoding.DecodeString(basic)
	// Not sure how to trigger this path
	if err != nil {
		return false, nil
	}

	parts := strings.SplitN(string(b), ":", 2)
	if len(parts) != 2 {
		SendLogoutCookie(c)
		return false, nil
	}

	username := parts[0]
	password := parts[1]
	hash := HashIt(username, password)

	// TODO this is ugly
	userUUID, err := config.LookupBasic(username, hash)
	if err != nil {
		SendLogoutCookie(c)
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
