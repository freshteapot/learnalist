package remind

import (
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/sirupsen/logrus"

	"github.com/labstack/echo/v4"
)

type RemindSpacedRepetitionService struct {
	remindRepo RemindSpacedRepetitionRepository
	logContext logrus.FieldLogger
}

func NewRemindSpacedRepetitionService(remindRepo RemindSpacedRepetitionRepository, log logrus.FieldLogger) RemindSpacedRepetitionService {
	s := RemindSpacedRepetitionService{
		remindRepo: remindRepo,
		logContext: log,
	}
	return s
}

func (s RemindSpacedRepetitionService) SetSpacedRepetition(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid

	defer c.Request().Body.Close()
	var input openapi.AppSettingsRemindV1
	json.NewDecoder(c.Request().Body).Decode(&input)

	if input.SpacedRepetition.PushEnabled < 0 || input.SpacedRepetition.PushEnabled > 1 {
		return c.JSON(http.StatusUnprocessableEntity, api.HTTPResponseMessage{
			Message: "push_enabled can only be 1 or 0",
		})
	}

	enabled := input.SpacedRepetition.PushEnabled
	err := s.remindRepo.SetPushEnabled(userUUID, enabled)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPErrorResponse)
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID:   userUUID,
		Kind:   EventApiRemindAppSettingsRemindV1,
		Data:   input,
		Action: event.ActionUpsert,
	})
	return c.JSON(http.StatusOK, input)
}
