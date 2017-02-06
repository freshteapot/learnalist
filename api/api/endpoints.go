package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo"
)

type (
	responseMessage struct {
		Message string `json:"message"`
	}
)

func (env *Env) GetRoot(c echo.Context) error {
	message := "1, 2, 3. Lets go!"
	response := &responseMessage{
		Message: message,
	}

	return c.JSON(http.StatusOK, response)
}

func (env *Env) GetListsByMe(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	alists, err := env.Datastore.GetListsBy(user.Uuid)
	if err != nil {
		message := fmt.Sprintf("Failed to find all lists.")
		response := new(responseMessage)
		response.Message = message
		return c.JSON(http.StatusBadRequest, *response)
	}
	return c.JSON(http.StatusOK, alists)
}

func (env *Env) GetListByUUID(c echo.Context) error {
	uuid := c.Param("uuid")
	alist, err := env.Datastore.GetAlist(uuid)
	if err != nil {
		message := fmt.Sprintf("Failed to find alist with uuid: %s", uuid)
		response := new(responseMessage)
		response.Message = message
		return c.JSON(http.StatusBadRequest, *response)
	}
	return c.JSON(http.StatusOK, *alist)
}

func (env *Env) PostAlist(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	playList := uuid.NewPlaylist(&user)
	uuid := playList.Uuid
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	aList := new(alist.Alist)
	err := aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		message := fmt.Sprintf("Your Json has a problem. %s", err)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusBadRequest, *response)
	}
	aList.Uuid = uuid
	aList.User = user
	env.Datastore.PostAlist(uuid, *aList)
	return c.JSON(http.StatusOK, *aList)
}

func (env *Env) PutAlist(c echo.Context) error {
	var err error
	var jsonBytes []byte
	// @todo do I not need to lock this down by logged in user?
	uuid := c.Param("uuid")
	defer c.Request().Body.Close()
	jsonBytes, _ = ioutil.ReadAll(c.Request().Body)

	aList := new(alist.Alist)
	err = aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		message := fmt.Sprintf("Your Json has a problem. %s", err)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusBadRequest, *response)
	}
	aList.Uuid = uuid

	err = env.Datastore.UpdateAlist(*aList)
	return c.JSON(http.StatusOK, *aList)
}

func (env *Env) RemoveAlist(c echo.Context) error {
	var message string
	uuid := c.Param("uuid")
	err := env.Datastore.RemoveAlist(uuid)
	response := &responseMessage{}

	message = fmt.Sprintf("List %s was removed.", uuid)
	if err != nil {
		message = fmt.Sprintf("Your Json has a problem. %s", err)
		response.Message = message
		return c.JSON(http.StatusBadRequest, *response)
	}
	response.Message = message
	return c.JSON(http.StatusOK, *response)
}

func (env *Env) PostRegister(c echo.Context) error {
	newUser, err := env.Datastore.InsertNewUser(c)
	if err != nil {
		return c.String(http.StatusBadRequest, "Bad json.")
	}
	return c.JSON(http.StatusOK, *newUser)
}
