package spaced_repetition

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func NewService(db *sqlx.DB) service {
	return service{
		db:   db,
		repo: NewSqliteRepository(db),
	}
}

func (s service) Endpoints(group *echo.Group) {
	group.GET("/next", s.GetNext)
	group.GET("/all", s.GetAll)
	group.DELETE("/:uuid", s.DeleteEntry)
	group.POST("/", s.SaveEntry)
	group.POST("/viewed", s.EntryViewed)
}

// SaveEntry Add entry for spaced based learning
func (s service) SaveEntry(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

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

	item := SpacedRepetitionEntry{
		UserUUID: user.Uuid,
		UUID:     entry.UUID(),
		Body:     entry.String(),
		WhenNext: entry.WhenNext(),
	}

	err := s.repo.SaveEntry(item)
	statusCode := http.StatusCreated
	if err != nil {
		if err != ErrSpacedRepetitionEntryExists {
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}
		statusCode = http.StatusOK
	}

	current, err := s.repo.GetEntry(item.UserUUID, item.UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	return c.JSON(statusCode, current)
}

// DeleteEntry Deletes a single entry based on the UUID
func (s service) DeleteEntry(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	UUID := c.Param("uuid")

	if UUID == "" {
		response := api.HttpResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	err := s.repo.DeleteEntry(user.Uuid, UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	return c.NoContent(http.StatusNoContent)
}

// GetNext Get next entry for spaced based learning
func (s service) GetNext(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	body, err := s.repo.GetNext(user.Uuid)

	if err != nil {
		if err == ErrNotFound {
			return c.NoContent(http.StatusNotFound)
		}

		if err == ErrFoundNotTime {
			return c.NoContent(http.StatusNoContent)
		}

		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	return c.JSON(http.StatusOK, body)
}

//GetAll Get all entries for spaced repetition learning
func (s service) GetAll(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

	items, err := s.repo.GetEntries(user.Uuid)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	return c.JSON(http.StatusOK, items)
}

func (s service) EntryViewed(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

	// Lookup uuid
	defer c.Request().Body.Close()

	var input HttpRequestViewed
	json.NewDecoder(c.Request().Body).Decode(&input)

	item := SpacedRepetitionEntry{}
	// TODO might need to update all time stamps to DATETIME as time.Time gets sad when stirng
	err := s.db.Get(&item, SQL_GET_NEXT, user.Uuid)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	// TODO could get this via the json_XXX functions in sqlite
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
	// save to db
	err = s.repo.UpdateEntry(item)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	current, err := s.repo.GetEntry(item.UserUUID, item.UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	return c.JSON(http.StatusOK, current)
}
