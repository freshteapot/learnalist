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
			uri                  = "/api/v1/api/v1/spaced-repetition/overtime"
			dripfeedHTTPResponse string
			inputFake, input     openapi.SpacedRepetitionOvertimeInputBase
		)

		BeforeEach(func() {
			dripfeedHTTPResponse = `{"dripfeed_uuid":"f29a45249551ae992a8edc6526ca7421094c8883","alist_uuid":"fake-list-123","user_uuid":"fake-user-123"}`
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

		It("Check filtering of supported lists", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
			aList := alist.Alist{}
			aList.Info.ListType = alist.Concept2
			listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			// Hardcoding to catch when I change it
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Kind not supported: v1,v2")
		})

		It("Issue looking upto see if dripfeed exists", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
			aList := alist.Alist{}
			aList.Uuid = input.AlistUuid
			aList.Info.ListType = alist.SimpleList

			dripfeedUUID := dripfeed.UUID(input.UserUuid, input.AlistUuid)
			listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, want)

			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("List already added, nothing more todo here", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
			aList := alist.Alist{}
			aList.Uuid = input.AlistUuid
			aList.Info.ListType = alist.SimpleList

			dripfeedUUID := dripfeed.UUID(input.UserUuid, input.AlistUuid)
			listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
			dripfeedRepo.On("Exists", dripfeedUUID).Return(true, nil)

			service.Create(c)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(dripfeedHTTPResponse))
		})

		When("We will add over time", func() {
			It("v1", func() {
				b, _ := json.Marshal(input)
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
				c = e.NewContext(req, rec)

				c.Set("loggedInUser", *loggedInUser)
				c.SetPath(uri)

				aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
				aList := alist.Alist{}
				aList.Uuid = input.AlistUuid
				aList.Info.ListType = alist.SimpleList
				aList.Data = make(alist.TypeV1, 0)
				aList.Data = append(aList.Data.(alist.TypeV1), "hello")

				dripfeedUUID := dripfeed.UUID(input.UserUuid, input.AlistUuid)
				listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
				dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)

				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiDripfeed))
					Expect(moment.Action).To(Equal(event.ActionCreated))

					b, _ := json.Marshal(moment.Data)
					var data dripfeed.EventDripfeedInputV1
					json.Unmarshal(b, &data)

					Expect(data.Info.AlistUUID).To(Equal(input.AlistUuid))
					Expect(data.Info.UserUUID).To(Equal(input.UserUuid))
					Expect(data.Info.Kind).To(Equal(alist.SimpleList))

					return true
				}))

				service.Create(c)

				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(dripfeedHTTPResponse))
			})

			When("v2", func() {
				It("show is not valid", func() {
					input := openapi.SpacedRepetitionOvertimeInputV2{
						AlistUuid: "fake-list-123",
						UserUuid:  userUUID,
						Settings: openapi.SpacedRepetitionOvertimeInputV2AllOfSettings{
							Show: "fake",
						},
					}
					b, _ := json.Marshal(input)
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
					c = e.NewContext(req, rec)

					c.Set("loggedInUser", *loggedInUser)
					c.SetPath(uri)

					aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
					aList := alist.Alist{}
					aList.Uuid = input.AlistUuid
					aList.Info.ListType = alist.FromToList
					aList.Data = make(alist.TypeV2, 0)

					dripfeedUUID := dripfeed.UUID(input.UserUuid, input.AlistUuid)
					listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
					dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)

					service.Create(c)

					Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
					testutils.CheckMessageResponseFromResponseRecorder(rec, "settings.show is not supported: from,to")
				})

				It("show is valid", func() {
					input := openapi.SpacedRepetitionOvertimeInputV2{
						AlistUuid: "fake-list-123",
						UserUuid:  userUUID,
						Settings: openapi.SpacedRepetitionOvertimeInputV2AllOfSettings{
							Show: "from",
						},
					}
					b, _ := json.Marshal(input)
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
					c = e.NewContext(req, rec)

					c.Set("loggedInUser", *loggedInUser)
					c.SetPath(uri)

					aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
					aList := alist.Alist{}
					aList.Uuid = input.AlistUuid
					aList.Info.ListType = alist.FromToList
					aList.Data = make(alist.TypeV2, 0)

					dripfeedUUID := dripfeed.UUID(input.UserUuid, input.AlistUuid)
					listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
					dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)

					eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.ApiDripfeed))
						Expect(moment.Action).To(Equal(event.ActionCreated))

						b, _ := json.Marshal(moment.Data)
						var data dripfeed.EventDripfeedInputV2
						json.Unmarshal(b, &data)

						Expect(data.Info.AlistUUID).To(Equal(input.AlistUuid))
						Expect(data.Info.UserUUID).To(Equal(input.UserUuid))
						Expect(data.Info.Kind).To(Equal(alist.FromToList))
						fmt.Println(data.Settings)
						return true
					}))

					service.Create(c)

					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(dripfeedHTTPResponse))
				})
			})
		})
	})
})
