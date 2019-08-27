package authenticate

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/labstack/echo/v4"
)

var LookUp func(loginUser LoginUser) (*uuid.User, error)

func SkipBasicAuth(c echo.Context) bool {
	url := c.Request().URL.Path
	method := c.Request().Method
	url = strings.TrimPrefix(url, "/api/v1")
	if url == "/" {
		return true
	}

	if url == "/version" && method == "GET" {
		return true
	}

	// Allow some more
	if strings.HasPrefix(c.Request().RemoteAddr, "127.0.0.1:") {
		method := c.Request().Method
		if url == "/register" && method == "POST" {
			return true
		}
	}
	return false
}

func ValidateBasicAuth(username string, password string, c echo.Context) (bool, error) {
	loginUser := &LoginUser{
		Username: username,
		Password: password,
	}
	user, err := LookUp(*loginUser)
	if err != nil {
		fmt.Println(err)
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
