package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

func (env *Env) GetListsByMe(c echo.Context) error {
	var alists []*alist.Alist
	user := c.Get("loggedInUser").(uuid.User)
	filterByLabels := c.QueryParam("labels")
	if filterByLabels == "" {
		alists = env.Datastore.GetListsByUser(user.Uuid)
	} else {
		alists = env.Datastore.GetListsByUserAndLabels(user.Uuid, filterByLabels)
	}

	return c.JSON(http.StatusOK, alists)
}

func (env *Env) GetListByUUID(c echo.Context) error {
	uuid := c.Param("uuid")
	alist, err := env.Datastore.GetAlist(uuid)
	if err != nil {
		message := fmt.Sprintf("Failed to find alist with uuid: %s", uuid)
		response := HttpResponseMessage{
			Message: message,
		}
		return c.JSON(http.StatusNotFound, response)
	}
	return c.JSON(http.StatusOK, *alist)
}

func (env *Env) SaveAlist(c echo.Context) error {
	var inputUuid string
	user := c.Get("loggedInUser").(uuid.User)
	method := c.Request().Method
	if method == http.MethodPost {
		playList := uuid.NewPlaylist(&user)
		inputUuid = playList.Uuid
	} else if method == http.MethodPut {
		inputUuid = c.Param("uuid")
	} else {
		response := HttpResponseMessage{
			Message: "This method is not supported.",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	aList := new(alist.Alist)
	err := aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		message := fmt.Sprintf("Your Json has a problem. %s", err)
		response := HttpResponseMessage{
			Message: message,
		}

		return c.JSON(http.StatusBadRequest, response)
	}

	aList.Uuid = inputUuid
	aList.User = user

	err = env.Datastore.SaveAlist(*aList)
	if err != nil {
		response := HttpResponseMessage{
			Message: err.Error(),
		}
		return c.JSON(http.StatusBadRequest, response)
	}
	return c.JSON(http.StatusOK, *aList)
}

func (env *Env) RemoveAlist(c echo.Context) error {
	var message string
	alist_uuid := c.Param("uuid")
	user := c.Get("loggedInUser").(uuid.User)
	err := env.Datastore.RemoveAlist(alist_uuid, user.Uuid)
	response := HttpResponseMessage{}

	message = fmt.Sprintf("List %s was removed.", alist_uuid)
	if err != nil {
		message = fmt.Sprintf("Your Json has a problem. %s", err)
		response.Message = message
		return c.JSON(http.StatusBadRequest, response)
	}
	response.Message = message
	return c.JSON(http.StatusOK, response)
}
