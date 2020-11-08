package challenge

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"

	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type ChallengeService struct {
	repo       ChallengeRepository
	acl        acl.AclChallenge
	logContext logrus.FieldLogger
}

func NewService(repo ChallengeRepository, acl acl.AclChallenge, log logrus.FieldLogger) ChallengeService {
	s := ChallengeService{
		repo:       repo,
		acl:        acl,
		logContext: log,
	}

	event.GetBus().Subscribe("challenge", func(entry event.Eventlog) {
		switch entry.Kind {
		case event.ApiUserDelete:
			s.removeUser(entry)
		case EventChallengeDone:
			s.eventChallengeDone(entry)
		case EventChallengeNewRecord:
			s.eventNotify(entry)
		}
	})
	return s
}

func (s ChallengeService) Challenges(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	lookupUserUUID := c.Param("userUUID")

	if lookupUserUUID != userUUID {
		response := api.HTTPResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	challenges, _ := s.repo.GetChallengesByUser(lookupUserUUID)
	return c.JSON(http.StatusOK, challenges)
}

func (s ChallengeService) Create(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid

	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	var challengeInput ChallengeInfo
	json.Unmarshal(jsonBytes, &challengeInput)

	if challengeInput.Description == "" {
		response := api.HTTPResponseMessage{
			Message: "Description cant be empty",
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	if challengeInput.Kind != KindPlankGroup {
		response := api.HTTPResponseMessage{
			Message: "Not valid kind",
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	UUID := guuid.New()
	challengeUUID := UUID.String()

	challengeInput.UUID = challengeUUID
	challengeInput.Created = ""
	challengeInput.Records = make([]ChallengePlankRecord, 0)
	challengeInput.Users = make([]ChallengePlankUser, 0)

	err := s.repo.Create(userUUID, challengeInput)

	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	_ = s.acl.MakeChallengePrivate(challengeUUID, userUUID)
	_ = s.acl.ShareChallengeWithPublic(challengeUUID)
	_ = s.acl.GrantUserChallengeWriteAccess(challengeUUID, userUUID)

	challenge, err := s.repo.Get(challengeUUID)

	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	return c.JSON(http.StatusCreated, challenge)
}

func (s ChallengeService) Join(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	challengeUUID := c.Param("uuid")
	_, err := s.repo.Get(challengeUUID)
	if err != nil {
		if err == ErrNotFound {
			return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
				Message: i18n.PlankRecordNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		})
	}

	allow, err := s.acl.HasUserChallengeWriteAccess(challengeUUID, userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		})
	}

	if allow {
		return c.NoContent(http.StatusOK)
	}

	_ = s.acl.GrantUserChallengeWriteAccess(challengeUUID, userUUID)

	_ = s.repo.Join(challengeUUID, userUUID)
	return c.NoContent(http.StatusOK)
}

func (s ChallengeService) Leave(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	challengeUUID := c.Param("uuid")

	_, err := s.repo.Get(challengeUUID)
	if err != nil {
		if err == ErrNotFound {
			return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
				Message: i18n.PlankRecordNotFound,
			})
		}

		return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		})
	}

	allow, err := s.acl.HasUserChallengeWriteAccess(challengeUUID, userUUID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		})
	}

	if !allow {
		return c.JSON(http.StatusForbidden, api.HTTPResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		})
	}

	_ = s.acl.RevokeUserChallengeWriteAccess(challengeUUID, userUUID)
	_ = s.repo.Leave(challengeUUID, userUUID)
	// Keep the records
	return c.NoContent(http.StatusOK)
}

func (s ChallengeService) Delete(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	challengeUUID := c.Param("uuid")

	allow, err := s.acl.HasUserChallengeOwnerAccess(challengeUUID, userUUID)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	if !allow {
		response := api.HTTPResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	_ = s.repo.Delete(challengeUUID)
	_ = s.acl.DeleteChallenge(challengeUUID)
	return c.NoContent(http.StatusOK)
}

func (s ChallengeService) Get(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	challengeUUID := c.Param("uuid")

	allow, err := s.acl.HasUserChallengeWriteAccess(challengeUUID, userUUID)
	if err != nil {
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

	challenge, err := s.repo.Get(challengeUUID)
	if err != nil {
		if err == ErrNotFound {
			return c.JSON(http.StatusNotFound, api.HTTPResponseMessage{
				Message: i18n.PlankRecordNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		})
	}

	return c.JSON(http.StatusOK, challenge)
}
