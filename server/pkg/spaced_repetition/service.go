package spaced_repetition

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

func NewService(db *sqlx.DB) SpacedRepetitionService {
	return SpacedRepetitionService{
		db:   db,
		repo: NewSqliteRepository(db),
	}
}

// SaveEntry Add entry for spaced based learning
func (s SpacedRepetitionService) SaveEntry(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

	defer c.Request().Body.Close()

	var temp interface{}
	json.NewDecoder(c.Request().Body).Decode(&temp)
	raw, _ := json.Marshal(temp)

	var what HTTPRequestInputKind
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
		Created:  entry.Created(),
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

	if statusCode == http.StatusCreated {
		event.GetBus().Publish(event.Eventlog{
			Kind: EventApiSpacedRepetition,
			Data: EventSpacedRepetition{
				Kind: EventKindNew,
				Data: item,
			},
		})

		// Baked the challenge system into the service
		// VS
		// UI needs more complexity
		challengeUUID := c.Request().Header.Get("challenge")
		if challengeUUID != "" {
			event.GetBus().Publish(event.Eventlog{
				Kind: challenge.EventChallengeDone,
				Data: challenge.EventChallengeDoneEntry{
					UUID:     challengeUUID,
					UserUUID: item.UserUUID,
					Data:     item.Body,
					Kind:     challenge.EventKindSpacedRepetition,
				},
			})
		}
	}

	return c.JSON(statusCode, current)
}

// DeleteEntry Deletes a single entry based on the UUID
func (s SpacedRepetitionService) DeleteEntry(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	UUID := c.Param("uuid")

	if UUID == "" {
		response := api.HTTPResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// TODO check if entry exsits
	err := s.repo.DeleteEntry(user.Uuid, UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	// This event fires, even if the entry doesnt exist
	event.GetBus().Publish(event.Eventlog{
		Kind: EventApiSpacedRepetition,
		Data: EventSpacedRepetition{
			Kind: EventKindDeleted,
			Data: SpacedRepetitionEntry{
				UUID:     UUID,
				UserUUID: user.Uuid,
			},
		},
	})

	return c.NoContent(http.StatusNoContent)
}

// GetNext Get next entry for spaced based learning
func (s SpacedRepetitionService) GetNext(c echo.Context) error {
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
func (s SpacedRepetitionService) GetAll(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

	items, err := s.repo.GetEntries(user.Uuid)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	return c.JSON(http.StatusOK, items)
}

func (s SpacedRepetitionService) EntryViewed(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

	// Lookup uuid
	defer c.Request().Body.Close()

	var input HTTPRequestViewed
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

	var what HTTPRequestInputKind
	json.Unmarshal([]byte(item.Body), &what)

	var entry ItemInput
	if what.Kind == "v1" {
		entry = V1FromDB(item.Body)
	}

	if what.Kind == "v2" {
		entry = V2FromDB(item.Body)
	}

	// TODO we do not protect actions

	// increment level
	// increment threshold
	// TODO change to const
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

	event.GetBus().Publish(event.Eventlog{
		Kind: EventApiSpacedRepetition,
		Data: EventSpacedRepetition{
			Kind:   EventKindViewed,
			Action: input.Action,
			Data:   item,
		},
	})

	current, err := s.repo.GetEntry(item.UserUUID, item.UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	return c.JSON(http.StatusOK, current)
}
