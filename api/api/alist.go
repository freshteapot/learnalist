package api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/labstack/echo/v4"
)

func (env *Env) GetListsByMe(c echo.Context) error {
	var alists []*alist.Alist
	user := c.Get("loggedInUser").(uuid.User)
	r := c.Request()
	params := r.URL.Query()
	if filterByLabels, ok := params["labels"]; ok {
		alists = env.Datastore.GetListsByUserAndLabels(user.Uuid, filterByLabels[0])
	} else {
		alists = env.Datastore.GetListsByUser(user.Uuid)
	}

	return c.JSON(http.StatusOK, alists)
}

func (env *Env) GetListByUUID(c echo.Context) error {
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	uuid := strings.TrimPrefix(r.URL.Path, "/alist/")
	if uuid == "" {
		response := HttpResponseMessage{
			Message: InputMissingListUuid,
		}
		return c.JSON(http.StatusNotFound, response)
	}
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
		// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
		r := c.Request()
		inputUuid = strings.TrimPrefix(r.URL.Path, "/alist/")
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

	statusCode := http.StatusOK
	if method == http.MethodPost {
		statusCode = http.StatusCreated
	}
	return c.JSON(statusCode, *aList)
}

func (env *Env) RemoveAlist(c echo.Context) error {
	var message string
	r := c.Request()
	// TODO Reference https://github.com/freshteapot/learnalist-api/issues/22
	alist_uuid := strings.TrimPrefix(r.URL.Path, "/alist/")

	user := c.Get("loggedInUser").(uuid.User)

	err := env.Datastore.RemoveAlist(alist_uuid, user.Uuid)
	response := HttpResponseMessage{}

	message = fmt.Sprintf("List %s was removed.", alist_uuid)
	if err != nil {
		response.Message = InternalServerErrorDeleteAlist
		return c.JSON(http.StatusInternalServerError, response)
	}
	response.Message = message
	return c.JSON(http.StatusOK, response)
}
