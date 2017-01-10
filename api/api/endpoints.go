package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist/api/alist"
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

func (env *Env) GetListsBy(c echo.Context) error {
	uuid := c.Param("uuid")
	alists, err := env.Datastore.GetListsBy(uuid)
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
	uuid := getUUID()

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	aList := new(alist.Alist)
	aList.Uuid = uuid
	err := aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		message := fmt.Sprintf("Your Json has a problem. %s", err)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusBadRequest, *response)
	}

	env.Datastore.PostAlist(uuid, *aList)
	return c.JSON(http.StatusOK, *aList)
}

func (env *Env) PutAlist(c echo.Context) error {
	var err error
	var jsonBytes []byte

	uuid := c.Param("uuid")
	defer c.Request().Body.Close()
	jsonBytes, _ = ioutil.ReadAll(c.Request().Body)

	aList := new(alist.Alist)
	aList.Uuid = uuid
	err = aList.UnmarshalJSON(jsonBytes)
	if err != nil {
		message := fmt.Sprintf("Your Json has a problem. %s", err)
		response := &responseMessage{
			Message: message,
		}
		return c.JSON(http.StatusBadRequest, *response)
	}

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
