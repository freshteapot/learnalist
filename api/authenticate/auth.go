package authenticate

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo"
)

var LookUp func(loginUser LoginUser) (*uuid.User, error)

func SkipBasicAuth(c echo.Context) bool {
	url := c.Request().URL.Path
	if url == "/" {
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

var basicAuth = "chris:chris"

func ValidateBasicAuth(username string, password string, c echo.Context) bool {
	loginUser := &LoginUser{
		Username: username,
		Password: password,
	}
	user, err := LookUp(*loginUser)
	if err != nil {
		fmt.Println(err)
		return false
	}

	c.Set("loggedInUser", *user)
	return true
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
	storedHash := fmt.Sprintf("A%d", hash)
	return storedHash, nil
}
