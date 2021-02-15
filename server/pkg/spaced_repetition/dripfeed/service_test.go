package dripfeed_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition/dripfeed"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Dripfeed Service API", func() {
	var (
		aclRepo         *mocks.Acl
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder
		service         dripfeed.DripfeedService
		dripfeedRepo    *mocks.DripfeedRepository
		listRepo        *mocks.DatastoreAlists
		loggedInUser    *uuid.User
		userUUID        string
		want            error
	)

	BeforeEach(func() {
		want = errors.New("want")
		loggedInUser = &uuid.User{
			Uuid: "fake-user-123",
		}
		userUUID = loggedInUser.Uuid

		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		e = echo.New()

		logger, _ = test.NewNullLogger()
		dripfeedRepo = &mocks.DripfeedRepository{}
		aclRepo = &mocks.Acl{}
		listRepo = &mocks.DatastoreAlists{}
		eventMessageBus.On("Subscribe", event.TopicMonolog, "dripfeedService", mock.Anything)

		service = dripfeed.NewService(dripfeedRepo, aclRepo, listRepo, logger)
	})

	When("Create", func() {
		var (
			uri              = "/api/v1/api/v1/spaced-repetition/overtime"
			inputFake, input openapi.SpacedRepetitionOvertimeInputBase
		)

		BeforeEach(func() {
			inputFake = openapi.SpacedRepetitionOvertimeInputBase{
				AlistUuid: "fake-list-123",
				UserUuid:  "fake-user-456",
			}
			input = openapi.SpacedRepetitionOvertimeInputBase{
				AlistUuid: "fake-list-123",
				UserUuid:  userUUID,
			}
		})

		It("User is not the one logged in", func() {
			b, _ := json.Marshal(inputFake)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)
			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "User doesnt match")
		})

		It("Failed to look up acl access", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(false, want)
			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorAclLookup)
		})

		It("Do not have access", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(false, nil)
			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.AclHttpAccessDeny)
		})

		It("Error looking up list", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
			listRepo.On("GetAlist", input.AlistUuid).Return(alist.Alist{}, want)

			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Error looking up list", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
			listRepo.On("GetAlist", input.AlistUuid).Return(alist.Alist{}, i18n.ErrorListNotFound)

			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			testutils.CheckMessageResponseFromResponseRecorder(rec, fmt.Sprintf(i18n.ApiAlistNotFound, input.AlistUuid))
		})
	})

})
