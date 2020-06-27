package spaced_repetition

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func NewService(db *sqlx.DB) service {
	return service{
		db: db,
	}
}

func (s service) Endpoints(group *echo.Group) {
	group.GET("/next", s.GetNext)
	group.GET("/list", s.GetAll)
	group.DELETE("/:uuid", s.DeleteItem)
	group.POST("/", s.SaveItem)
}

func (s service) SaveItem(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	fmt.Println(user)

	defer c.Request().Body.Close()

	var input HttpRequestInput
	json.NewDecoder(c.Request().Body).Decode(&input)

	// Set level
	input.Settings.Level = Level_0
	b, _ := json.Marshal(input.Data)
	hash := fmt.Sprintf("%x", sha1.Sum(b))

	// Based on level
	whenNext := time.Now().Add(time.Hour * 1).UTC()
	input.Settings.WhenNext = whenNext.Format(time.RFC3339)
	b, _ = json.Marshal(input)

	item := SpacedRepetitionItem{
		UserUUID: user.Uuid,
		UUID:     hash,
		Body:     string(b),
		WhenNext: whenNext,
	}

	_, err := s.db.Exec(SQL_SAVE_ITEM, item.UUID, item.Body, item.UserUUID, item.WhenNext)
	fmt.Println(err)
	// Write to db
	/*
		uuid = hash of data
		user_uuid,
		body,
		when

		unique index = (uuid, user_hash)
	*/
	response := api.HttpResponseMessage{
		Message: "TODO",
	}
	return c.JSON(http.StatusOK, response)
}

func (s service) DeleteItem(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	UUID := c.Param("uuid")
	fmt.Println(user)

	if UUID == "" {
		response := api.HttpResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusNotFound, response)
	}

	_, err := s.db.Exec(SQL_DELETE_ITEM, UUID, user.Uuid)
	fmt.Println(err)
	response := api.HttpResponseMessage{
		Message: "TODO",
	}
	return c.JSON(http.StatusOK, response)
}

func (s service) GetNext(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	fmt.Println(user)
	response := api.HttpResponseMessage{
		Message: "TODO",
	}
	return c.JSON(http.StatusOK, response)
}

func (s service) GetAll(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	fmt.Println(user)
	response := api.HttpResponseMessage{
		Message: "TODO",
	}
	return c.JSON(http.StatusOK, response)
}
