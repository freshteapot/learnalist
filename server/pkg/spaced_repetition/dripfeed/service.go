package dripfeed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/i18n"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// @openapi.path.tag: spacedRepetition
func NewService(repo DripfeedRepository, aclRepo acl.AclReaderList, listRepo alist.DatastoreAlists, logContext logrus.FieldLogger) DripfeedService {
	s := DripfeedService{
		repo:       repo,
		aclRepo:    aclRepo,
		listRepo:   listRepo,
		logContext: logContext,
	}

	event.GetBus().Subscribe(event.TopicMonolog, "dripfeedService", s.OnEvent)
	return s
}

// @event.emit: event.ApiDripfeed
func (s DripfeedService) Create(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	logContext := s.logContext

	defer c.Request().Body.Close()

	var temp interface{}
	json.NewDecoder(c.Request().Body).Decode(&temp)
	raw, _ := json.Marshal(temp)

	var input openapi.SpacedRepetitionOvertimeInputBase
	json.Unmarshal(raw, &input)

	if input.UserUuid != loggedInUser.Uuid {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "User doesnt match",
		})
	}

	allow, err := s.aclRepo.HasUserListReadAccess(input.AlistUuid, input.UserUuid)

	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "broken-state",
			"input": input,
			"error": err,
		}).Error("s.aclRepo.HasUserListReadAccess")

		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	if !allow {
		response := api.HTTPResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	aList, err := s.listRepo.GetAlist(input.AlistUuid)
	if err != nil {
		if err == i18n.ErrorListNotFound {
			logContext.WithFields(logrus.Fields{
				"event":      "broken-state",
				"alist_uuid": input.AlistUuid,
			}).Error("List not found, but has acl access")

			message := fmt.Sprintf(i18n.ApiAlistNotFound, input.AlistUuid)
			response := api.HTTPResponseMessage{
				Message: message,
			}
			return c.JSON(http.StatusNotFound, response)
		}
		// When the db fails to lookup, maybe we should actually be crashing.
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	allowedKinds := []string{alist.SimpleList, alist.FromToList}
	if !utils.StringArrayContains(allowedKinds, aList.Info.ListType) {
		response := api.HTTPResponseMessage{
			Message: fmt.Sprintf("Kind not supported: %s", strings.Join(allowedKinds, ",")),
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}
	// TODO maybe move this further up
	dripfeedUUID := UUID(input.UserUuid, input.AlistUuid)
	exists, err := s.repo.Exists(dripfeedUUID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "broken-state",
			"input": input,
			"error": err,
		}).Error("s.repo.Exists")

		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	dripfeedResponse := openapi.SpacedRepetitionOvertimeInfo{
		UserUuid:     loggedInUser.Uuid,
		AlistUuid:    aList.Uuid,
		DripfeedUuid: dripfeedUUID,
	}

	if exists {
		return c.JSON(http.StatusOK, dripfeedResponse)
	}

	info := EventDripfeedInputBase{
		UserUUID:  loggedInUser.Uuid,
		AlistUUID: aList.Uuid,
		Kind:      aList.Info.ListType,
	}

	var data interface{}
	switch info.Kind {
	case alist.SimpleList:
		data = EventDripfeedInputV1{
			Info: info,
			Data: aList.Data.(alist.TypeV1),
		}
	case alist.FromToList:
		var extra openapi.SpacedRepetitionOvertimeInputV2
		json.Unmarshal(raw, &extra)
		data = EventDripfeedInputV2{
			Info:     info,
			Settings: extra.Settings,
			Data:     aList.Data.(alist.TypeV2),
		}
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID:   dripfeedUUID,
		Kind:   event.ApiDripfeed,
		Data:   data,
		Action: event.ActionCreated,
	})

	return c.JSON(http.StatusOK, dripfeedResponse)
}

// @event.emit: event.ApiDripfeed
func (s DripfeedService) Delete(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid

	logContext := s.logContext.WithFields(logrus.Fields{
		"entry":     "delete",
		"user_uuid": userUUID,
	})

	defer c.Request().Body.Close()
	var input openapi.SpacedRepetitionOvertimeInputBase
	json.NewDecoder(c.Request().Body).Decode(&input)

	if input.UserUuid != loggedInUser.Uuid {
		logContext.WithFields(logrus.Fields{
			"error": "user-match",
			"input": input,
		}).Error("input")

		return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
			Message: "User doesnt match",
		})
	}

	dripfeedUUID := UUID(input.UserUuid, input.AlistUuid)

	logContext = logContext.WithFields(logrus.Fields{
		"input":         input,
		"dripfeed_uuid": dripfeedUUID,
	})

	exists, err := s.repo.Exists(dripfeedUUID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "broken-state",
			"error": err,
		}).Error("s.repo.Exists")

		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	if !exists {
		return c.JSON(http.StatusOK, api.HTTPResponseMessage{
			Message: "List removed",
		})
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID: dripfeedUUID,
		Kind: event.ApiDripfeed,
		Data: EventDripfeedDelete{
			DripfeedUUID: dripfeedUUID,
			UserUUID:     userUUID,
		},
		Action: event.ActionDeleted,
	})

	return c.JSON(http.StatusOK, api.HTTPResponseMessage{
		Message: "List removed",
	})
}

func (s DripfeedService) ListActive(c echo.Context) error {
	userUUID := c.Get("loggedInUser").(uuid.User).Uuid
	alistUUID := c.Param("alistUUID")

	dripfeedUUID := UUID(userUUID, alistUUID)

	logContext := s.logContext.WithFields(logrus.Fields{
		"entry":         "list-active",
		"user_uuid":     userUUID,
		"alist_uuid":    alistUUID,
		"dripfeed_uuid": dripfeedUUID,
	})

	exists, err := s.repo.Exists(dripfeedUUID)
	if err != nil {
		logContext.WithFields(logrus.Fields{
			"event": "broken-state",
			"error": err,
		}).Error("s.repo.Exists")
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	if !exists {
		return c.NoContent(http.StatusNotFound)
	}
	return c.NoContent(http.StatusOK)
}
