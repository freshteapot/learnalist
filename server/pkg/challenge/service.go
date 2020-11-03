package challenge

import (
	"encoding/json"
	"fmt"
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
		fmt.Println("TODO: take from slack events")

		switch entry.Kind {
		case event.ApiUserDelete:
			b, err := json.Marshal(entry.Data)
			if err != nil {
				return
			}

			var moment event.EventUser
			json.Unmarshal(b, &moment)
			s.repo.DeleteUser(moment.UUID)
			fmt.Println("delete user")
			return
		case EventChallengeDone:
			var moment EventChallengeDoneEntry
			b, _ := json.Marshal(entry.Data)
			json.Unmarshal(b, &moment)

			challengeUUID := moment.UUID
			if moment.Kind != EventKindPlank {
				s.logContext.WithFields(logrus.Fields{
					"kind":           moment.Kind,
					"challenge_uuid": challengeUUID,
					"user_uuid":      moment.UserUUID,
				}).Info("kind not supported")
				return
			}

			b, _ = json.Marshal(moment.Data)
			var record ChallengePlankRecord
			json.Unmarshal(b, &record)
			fmt.Printf("Write to db kind: %s, challenge:%s, user:%s, record:%s\n",
				moment.Kind,
				challengeUUID,
				moment.UserUUID,
				record.UUID,
			)
			// Do I tightly couple?
			// Why not the whole thing is tightly coupled
			// save plank
			// TODO how do I know when its deleted?
			// delete plank by userUUID
			err := s.repo.AddRecord(challengeUUID, record.UUID, moment.UserUUID)
			if err != nil {
				fmt.Println(err)
			}
			return
		default:
			return
		}

	})
	return s
}

func (s ChallengeService) Challenges(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	lookupUserUUID := c.Param("userUUID")
	if lookupUserUUID == "" {
		response := api.HTTPResponseMessage{
			Message: "Missing userUUID",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

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

	if challengeInput.Kind != "plank-group" {
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
	challengeInput.Users = []ChallengePlankUsers{
		{
			UserUUID: userUUID,
			Name:     "TODO - creator",
		},
	}

	err := s.repo.Create(userUUID, challengeInput)

	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	s.acl.MakeChallengePrivate(challengeUUID, userUUID)
	s.acl.ShareChallengeWithPublic(challengeUUID)
	s.acl.GrantUserChallengeWriteAccess(challengeUUID, userUUID)
	// Add user to the list
	challenge, err := s.repo.Get(challengeUUID)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		}
		return c.JSON(http.StatusInternalServerError, response)
	}

	// Let the data be challenge specific
	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: create challenge for %s", userUUID),
	}
	fmt.Println(response)
	return c.JSON(http.StatusOK, challenge)
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
			Message: i18n.InternalServerErrorAclLookup,
		})
	}

	allow, err := s.acl.HasUserChallengeWriteAccess(challengeUUID, userUUID)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorFunny,
		}
		return c.JSON(http.StatusInternalServerError, response)
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
			Message: i18n.InternalServerErrorAclLookup,
		})
	}

	allow, err := s.acl.HasUserChallengeWriteAccess(challengeUUID, userUUID)
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
