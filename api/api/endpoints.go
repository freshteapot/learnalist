package api

import (
	"fmt"
	"net/http"

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
	fmt.Println("here2")
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
	fmt.Println(uuid)
	fmt.Println(len(uuid))
	message := fmt.Sprintf("I want to upload alist with uuid: %s", uuid)
	response := &responseMessage{
		Message: message,
	}
	return c.JSON(http.StatusOK, response)
}

func (env *Env) PutAlist(c echo.Context) error {
	uuid := c.Param("uuid")
	message := fmt.Sprintf("I want to alter alist with uuid: %s", uuid)
	response := &responseMessage{
		Message: message,
	}
	return c.JSON(http.StatusOK, response)
}

func (env *Env) PatchAlist(c echo.Context) error {
	uuid := c.Param("uuid")
	message := fmt.Sprintf("I want to alter alist with uuid: %s", uuid)
	response := &responseMessage{
		Message: message,
	}
	return c.JSON(http.StatusOK, response)
}
