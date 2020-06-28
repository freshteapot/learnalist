package spaced_repetition

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
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
	group.GET("/list", s.GetAll) // I wonder if list or all is better
	group.GET("/all", s.GetAll)
	group.DELETE("/:uuid", s.DeleteItem)
	group.POST("/", s.SaveItem)
	group.POST("/viewed", s.ItemViewed)
}

func (s service) SaveItem(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	fmt.Println(user)

	defer c.Request().Body.Close()

	var temp interface{}
	json.NewDecoder(c.Request().Body).Decode(&temp)
	raw, _ := json.Marshal(temp)

	var what HttpRequestInputKind
	json.Unmarshal(raw, &what)
	if !utils.StringArrayContains([]string{"v1", "v2"}, what.Kind) {
		return c.NoContent(http.StatusBadRequest)
	}

	var entry ItemInput

	if what.Kind == "v1" {
		entry = V1FromPOST(raw)
	}

	if what.Kind == "v2" {
		entry = V2FromPOST(raw)
	}

	item := SpacedRepetitionItem{
		UserUUID: user.Uuid,
		UUID:     entry.UUID(),
		Body:     entry.String(),
		WhenNext: entry.WhenNext(),
	}

	_, err := s.db.Exec(SQL_SAVE_ITEM, item.UUID, item.Body, item.UserUUID, item.WhenNext, item.Body, item.WhenNext)
	fmt.Println(err)

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
	item := SpacedRepetitionItem{}
	// TODO might need to update all time stamps to DATETIME as time.Time gets sad when stirng
	err := s.db.Get(&item, SQL_GET_NEXT, user.Uuid)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNoContent)
		}

		fmt.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if !time.Now().UTC().After(item.WhenNext) {
		return c.NoContent(http.StatusNoContent)
	}

	var body interface{}
	json.Unmarshal([]byte(item.Body), &body)
	return c.JSON(http.StatusOK, body)
}

func (s service) GetAll(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	items := make([]interface{}, 0)
	dbItems := make([]string, 0)
	err := s.db.Select(&dbItems, SQL_GET_ALL, user.Uuid)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNoContent)
		}

		fmt.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	for _, item := range dbItems {
		var body interface{}
		json.Unmarshal([]byte(item), &body)
		items = append(items, body)
	}

	return c.JSON(http.StatusOK, items)
}

func (s service) ItemViewed(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

	// Lookup uuid

	defer c.Request().Body.Close()

	var input HttpRequestViewed
	json.NewDecoder(c.Request().Body).Decode(&input)

	item := SpacedRepetitionItem{}
	// TODO might need to update all time stamps to DATETIME as time.Time gets sad when stirng
	err := s.db.Get(&item, SQL_GET_NEXT, user.Uuid)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNotFound)
		}

		fmt.Println(err)
		return c.NoContent(http.StatusInternalServerError)
	}

	// hmm maybe add kind to the table

	var what HttpRequestInputKind
	json.Unmarshal([]byte(item.Body), &what)

	var entry ItemInput
	if what.Kind == "v1" {
		entry = V1FromDB(item.Body)
	}

	if what.Kind == "v2" {
		entry = V2FromDB(item.Body)
	}

	// increment level
	// increment threshold
	if input.Action == "incr" {
		entry.IncrThreshold()
	}

	if input.Action == "decr" {
		entry.DecrThreshold()
	}

	item.Body = entry.String()
	item.WhenNext = entry.WhenNext()
	fmt.Println(item.WhenNext)
	// save to db
	_, err = s.db.Exec(SQL_SAVE_ITEM, item.UUID, item.Body, item.UserUUID, item.WhenNext, item.Body, item.WhenNext)
	fmt.Println(err)

	response := api.HttpResponseMessage{
		Message: "TODO",
	}
	return c.JSON(http.StatusOK, response)
}

// This is the start of the "magic"
// Linking the data to active users
// Make it work
func (s service) CheckForNewItems() {
	// Get for all users?
	//
	fmt.Println("Check for new items")
}

//var Events := make(chan *sse.Event)
