package plank

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

type PlankService struct {
	repo       PlankRepository
	logContext logrus.FieldLogger
}

// @openapi.path.tag: plank
func NewService(repo PlankRepository, log logrus.FieldLogger) PlankService {
	s := PlankService{
		repo:       repo,
		logContext: log,
	}

	event.GetBus().Subscribe(event.TopicMonolog, "plank", s.monologSubscribe)
	return s
}

func (s PlankService) History(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	history, err := s.repo.History(user.Uuid)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}
	return c.JSON(http.StatusOK, history)
}

// RecordPlank Document the plank
// @event.emit: event.ApiPlank
// @event.emit: challenge.EventChallengeDone
func (s PlankService) RecordPlank(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)

	defer c.Request().Body.Close()

	var input openapi.Plank
	json.NewDecoder(c.Request().Body).Decode(&input)

	// Set to empty, to make sure the hash is the data not the hash
	// Could one day let the user do it, and confirm hash = hash.
	input.Uuid = ""
	b, _ := json.Marshal(input)
	hash := fmt.Sprintf("%x", sha1.Sum(b))
	input.Uuid = hash
	created := time.Unix(0, int64(input.BeginningTime)*int64(1000000))
	// TODO add validation

	item := Entry{
		UserUUID: user.Uuid,
		UUID:     hash,
		Body:     input,
		Created:  created.UTC(),
	}

	err := s.repo.SaveEntry(item)
	actuallySaved := true
	if err != nil {
		if err != ErrEntryExists {
			return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
		}
		actuallySaved = false
	}

	// If it was already in the system, return
	if !actuallySaved {
		return c.JSON(http.StatusOK, input)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiPlank,
		Data: event.EventPlank{
			Action:   event.ActionNew,
			UserUUID: item.UserUUID,
			Data:     item.Body,
		},
	})

	// Baked the challenge system into the service
	// VS
	// UI needs more complexity

	// Send event if challenge
	challengeUUID := c.Request().Header.Get("x-challenge")
	if challengeUUID != "" {
		event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
			Kind: challenge.EventChallengeDone,
			Data: challenge.EventChallengeDoneEntry{
				UUID:     challengeUUID,
				UserUUID: item.UserUUID,
				Data:     item.Body,
				Kind:     challenge.EventKindPlank,
			},
		})
	}

	return c.JSON(http.StatusCreated, input)
}

// DeletePlankRecord Deletes a single entry based on the UUID
// @event.emit: event.ApiPlank
func (s PlankService) DeletePlankRecord(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	UUID := c.Param("uuid")

	if UUID == "" {
		response := api.HTTPResponseMessage{
			Message: i18n.InputMissingListUuid,
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	record, err := s.repo.GetEntry(UUID, user.Uuid)
	if err != nil {
		if err == utils.ErrNotFound {
			return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
				Message: i18n.PlankRecordNotFound,
			})
		}
		s.logContext.WithFields(logrus.Fields{
			"event":     "delete.record.lookup",
			"uuid":      UUID,
			"user_uuid": user.Uuid,
		}).Error("repo")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	err = s.repo.DeleteEntry(UUID, user.Uuid)
	if err != nil {
		s.logContext.WithFields(logrus.Fields{
			"event":     "api.delete.record.remove",
			"uuid":      UUID,
			"user_uuid": user.Uuid,
		}).Error("repo")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: event.ApiPlank,
		Data: event.EventPlank{
			Action:   event.ActionDeleted,
			UserUUID: user.Uuid,
			Data:     record,
		},
	})

	return c.NoContent(http.StatusNoContent)
}

func (s PlankService) monologSubscribe(entry event.Eventlog) {
	if entry.Kind != event.ApiUserDelete {
		return
	}
	userUUID := entry.UUID
	s.repo.DeleteEntriesByUser(userUUID)
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": userUUID,
		"event":     event.UserDeleted,
	}).Info("entries removed")
}
