package challenge

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"

	guuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type ChallengeService struct {
	repo                                ChallengeRepository
	challengePushNotificationRepository ChallengePushNotificationRepository
	acl                                 acl.AclChallenge
	logContext                          logrus.FieldLogger
}

func NewService(repo ChallengeRepository, challengePushNotificationRepository ChallengePushNotificationRepository, acl acl.AclChallenge, log logrus.FieldLogger) ChallengeService {
	s := ChallengeService{
		repo:                                repo,
		challengePushNotificationRepository: challengePushNotificationRepository,
		acl:                                 acl,
		logContext:                          log,
	}

	event.GetBus().Subscribe(event.TopicMonolog, "challenge", s.OnEvent)
	return s
}

func (s ChallengeService) Challenges(c echo.Context) error {
	user := c.Get("loggedInUser").(uuid.User)
	userUUID := user.Uuid
	lookupUserUUID := c.Param("userUUID")
	filterByKind := c.QueryParam("kind")

	if lookupUserUUID != userUUID {
		response := api.HTTPResponseMessage{
			Message: i18n.AclHttpAccessDeny,
		}
		return c.JSON(http.StatusForbidden, response)
	}

	allowed := []string{"", KindPlankGroup}
	if !utils.StringArrayContains(allowed, filterByKind) {
		response := api.HTTPResponseMessage{
			Message: "Not valid kind",
		}
		return c.JSON(http.StatusUnprocessableEntity, response)
	}

	challenges, _ := s.repo.GetChallengesByUser(lookupUserUUID, filterByKind)
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

	allowed := ChallengeKinds
	if !utils.StringArrayContains(allowed, challengeInput.Kind) {
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

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: EventChallengeCreated,
		Data: event.EventKV{
			UUID: challengeUUID,
			Data: challenge,
		},
	})

	s.updateStaticSite(challenge, false, event.ActionCreated)
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
				Message: i18n.ChallengeNotFound,
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
	// I am not sure I need this
	// Or I need to move the above logic into it
	//_ = s.repo.Join(challengeUUID, userUUID)

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: EventChallengeJoined,
		Data: event.EventKV{
			UUID: challengeUUID,
			Data: ChallengeJoined{
				UUID:     challengeUUID,
				UserUUID: userUUID,
			},
		},
	})

	s.updateStaticSite(ChallengeInfo{UUID: challengeUUID}, true, event.ActionUpdated)
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
				Message: i18n.ChallengeNotFound,
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
	// I am not sure I need this
	// Or I need to move the above logic into it
	//_ = s.repo.Leave(challengeUUID, userUUID)

	// If I want to build a cache
	// listen to changes to displayName
	// listen to addrecord
	// listen to join
	// listen to leave

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		Kind: EventChallengeLeft,
		Data: event.EventKV{
			UUID: challengeUUID,
			Data: ChallengeLeft{
				UUID:     challengeUUID,
				UserUUID: userUUID,
			},
		},
	})
	s.updateStaticSite(ChallengeInfo{UUID: challengeUUID}, true, event.ActionUpdated)
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

	// TODO https://github.com/freshteapot/learnalist-api/issues/175
	s.updateStaticSite(ChallengeInfo{UUID: challengeUUID}, false, event.ActionDeleted)
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
				Message: i18n.ChallengeNotFound,
			})
		}
		return c.JSON(http.StatusInternalServerError, api.HTTPResponseMessage{
			Message: i18n.InternalServerErrorAclLookup,
		})
	}

	return c.JSON(http.StatusOK, challenge)
}

func (s ChallengeService) updateStaticSite(challenge ChallengeInfo, lookup bool, action string) {
	var err error
	// Known issue: when a user updates their display name. This will get out of sync.
	if lookup {
		uuid := challenge.UUID
		challenge, err = s.repo.Get(uuid)
		if err != nil {
			s.logContext.WithFields(logrus.Fields{
				"event":          "sync-challenge-to-static-site",
				"error":          err,
				"challenge_uuid": uuid,
			}).Error("challenge lookup failed, possibly db issue")
			return
		}
	}

	event.GetBus().Publish(event.TopicStaticSite, event.Eventlog{
		Kind: event.ChangesetChallenge,
		Data: challenge,
	})
}
