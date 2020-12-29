package spaced_repetition

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func NewService(repo SpacedRepetitionRepository, logContext logrus.FieldLogger) SpacedRepetitionService {
	s := SpacedRepetitionService{
		repo:       repo,
		logContext: logContext,
	}

	event.GetBus().Subscribe(event.TopicMonolog, "spacedRepetitionService", s.OnEvent)
	return s
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

	allowedKinds := []string{alist.SimpleList, alist.FromToList}
	if !utils.StringArrayContains(allowedKinds, what.Kind) {
		response := api.HTTPResponseMessage{
			Message: fmt.Sprintf("Kind not supported: %s", strings.Join(allowedKinds, ",")),
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	var entry ItemInput

	switch what.Kind {
	case alist.SimpleList:
		entry = V1FromPOST(raw)
	case alist.FromToList:
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

	var current interface{}
	json.Unmarshal([]byte(entry.String()), &current)

	if statusCode == http.StatusOK {
		return c.JSON(statusCode, current)
	}

	// The entry is a new
	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiSpacedRepetition,
		Data: EventSpacedRepetition{
			Kind: EventKindNew,
			Data: item,
		},
	})

	// Baked the challenge system into the service
	// VS
	// UI needs more complexity
	challengeUUID := c.Request().Header.Get("x-challenge")
	if challengeUUID != "" {
		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: challenge.EventChallengeDone,
			Data: challenge.EventChallengeDoneEntry{
				UUID:     challengeUUID,
				UserUUID: item.UserUUID,
				Data:     item.Body,
				Kind:     challenge.EventKindSpacedRepetition,
			},
		})
	}

	return c.JSON(statusCode, current)
}

// DeleteEntry Deletes a single entry based on the UUID
func (s SpacedRepetitionService) DeleteEntry(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	UUID := c.Param("uuid")
	userUUID := user.Uuid

	if UUID == "" {
		response := api.HTTPResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// Confirm the entry exists
	_, err := s.repo.GetEntry(userUUID, UUID)
	if err != nil {
		if err == utils.ErrNotFound {
			return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
				Message: "Entry not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	err = s.repo.DeleteEntry(userUUID, UUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiSpacedRepetition,
		Data: EventSpacedRepetition{
			Kind: EventKindDeleted,
			Data: SpacedRepetitionEntry{
				UUID:     UUID,
				UserUUID: userUUID,
			},
		},
	})

	return c.NoContent(http.StatusNoContent)
}

// GetNext Get next entry for spaced based learning
func (s SpacedRepetitionService) GetNext(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	body, err := CheckNext(s.repo.GetNext(user.Uuid))

	if err != nil {
		if err == utils.ErrNotFound {
			return c.NoContent(http.StatusNotFound)
		}

		if err == ErrFoundNotTime {
			return c.NoContent(http.StatusNoContent)
		}

		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	return c.JSON(http.StatusOK, body)
}

// GetAll Get all entries for spaced repetition learning
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

	var input openapi.SpacedRepetitionEntryViewed
	json.NewDecoder(c.Request().Body).Decode(&input)

	allowed := []string{ActionIncrement, ActionDecrement}
	if !utils.StringArrayContains(allowed, input.Action) {
		response := api.HTTPResponseMessage{
			Message: fmt.Sprintf("Action not supported: %s", strings.Join(allowed, ",")),
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	// TODO might need to update all time stamps to DATETIME as time.Time gets sad when stirng
	// TODO if I use GetEntry, then this is more generic and not locked to GetNext only
	item, err := s.repo.GetNext(user.Uuid)
	if err != nil {
		if err == utils.ErrNotFound {
			return c.NoContent(http.StatusNotFound)
		}

		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if input.Uuid != item.UUID {
		return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
			Message: "Input uuid is not the uuid of what is next",
		})
	}

	// At this point, we are assuming the list is valid
	var what HTTPRequestInputKind
	json.Unmarshal([]byte(item.Body), &what)

	var entry ItemInput

	switch what.Kind {
	case alist.SimpleList:
		entry = V1FromDB(item.Body)
	case alist.FromToList:
		entry = V2FromDB(item.Body)
	}

	// Based on the action, bubbles up when the entry will be scheduled for next viewing.
	switch input.Action {
	case ActionIncrement:
		entry.IncrThreshold()
	case ActionDecrement:
		entry.DecrThreshold()
	}

	item.Body = entry.String()
	item.WhenNext = entry.WhenNext()
	// save to db
	err = s.repo.UpdateEntry(item)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiSpacedRepetition,
		Data: EventSpacedRepetition{
			Kind:   EventKindViewed,
			Action: input.Action,
			Data:   item,
		},
	})

	var current interface{}
	json.Unmarshal([]byte(item.Body), &current)
	return c.JSON(http.StatusOK, current)
}
