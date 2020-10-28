package challenge

import (
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type ChallengeService struct {
	repo       ChallengeRepository
	logContext logrus.FieldLogger
}

func NewService(repo ChallengeRepository, log logrus.FieldLogger) ChallengeService {
	s := ChallengeService{
		repo:       repo,
		logContext: log,
	}

	event.GetBus().Subscribe("challenge", func(entry event.Eventlog) {
		fmt.Println("TODO: take from slack events")
	})
	return s
}

func (s ChallengeService) Challenges(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	lookupUserUUID := c.Param("userUUID")
	fmt.Println("user %s %s", userUUID, lookupUserUUID)
	if lookupUserUUID == "" {
		response := api.HTTPResponseMessage{
			Message: "Missing userUUID",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// TODO check access
	// TODO userUUID = UUID of the challenge/:userUUID
	// Will need system lists for this approach
	response := api.HTTPResponseMessage{
		Message: "TODO: Get challenges based on the userUUID",
	}
	return c.JSON(http.StatusOK, response)
}

func (s ChallengeService) Create(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: create challenge for %s", userUUID),
	}
	return c.JSON(http.StatusOK, response)
}

func (s ChallengeService) Join(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	challengeUUID := c.Param("uuid")

	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: user %s wants to join challenge %s", userUUID, challengeUUID),
	}
	return c.JSON(http.StatusOK, response)
}

func (s ChallengeService) Leave(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	challengeUUID := c.Param("uuid")

	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: user %s wants to leave challenge %s", userUUID, challengeUUID),
	}
	return c.JSON(http.StatusOK, response)
}

func (s ChallengeService) Get(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	challengeUUID := c.Param("uuid")

	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: Get challenge %s, for user %s", challengeUUID, userUUID),
	}
	return c.JSON(http.StatusOK, response)
}
