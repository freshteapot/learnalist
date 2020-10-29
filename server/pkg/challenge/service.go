package challenge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
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
	aListRepo  alist.DatastoreAlists
	acl        acl.Acl // Change to AclChallenge
	logContext logrus.FieldLogger
}

func NewService(repo ChallengeRepository, acl acl.Acl, log logrus.FieldLogger) ChallengeService {
	s := ChallengeService{
		repo:       repo,
		acl:        acl,
		logContext: log,
	}

	event.GetBus().Subscribe("challenge", func(entry event.Eventlog) {
		fmt.Println("TODO: take from slack events")

		if entry.Kind != EventChallengeDone {
			return
		}

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
		// delete plank by userUUID
		return

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

	var body ChallengeBody
	json.Unmarshal(jsonBytes, &body)

	if body.Description == "" {
		response := api.HTTPResponseMessage{
			Message: "Description cant be empty",
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	if body.Kind != "plank-group" {
		response := api.HTTPResponseMessage{
			Message: "Not valid kind",
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	UUID := guuid.New()
	challengeUUID := UUID.String()

	body.UUID = challengeUUID
	body.Created = ""

	b, _ := json.Marshal(body)
	err := s.repo.Create(ChallengeEntry{
		UUID:     challengeUUID,
		UserUUID: userUUID,
		Body:     string(b),
	})

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

	s.acl.GrantUserChallengeWriteAccess(challengeUUID, userUUID)

	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: user %s wants to join challenge %s", userUUID, challengeUUID),
	}
	return c.JSON(http.StatusOK, response)
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
	s.acl.RevokeUserChallengeWriteAccess(challengeUUID, userUUID)
	// TODO Keep records?
	// Soft delete?
	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: user %s wants to leave challenge %s", userUUID, challengeUUID),
	}
	return c.JSON(http.StatusOK, response)
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

	// Why not use list_type and then filter?
	response := api.HTTPResponseMessage{
		Message: fmt.Sprintf("TODO: Get challenge %s, for user %s", challengeUUID, userUUID),
	}
	fmt.Println(response.Message)
	return c.JSON(http.StatusOK, challenge)
}
