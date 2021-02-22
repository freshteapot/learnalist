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
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder

		service      dripfeed.DripfeedService
		dripfeedRepo *mocks.DripfeedRepository
		listRepo     *mocks.DatastoreAlists
		aclRepo      *mocks.Acl

		loggedInUser *uuid.User
		want         error

		userUUID     string
		alistUUID    string
		dripfeedUUID string
	)

	BeforeEach(func() {
		want = errors.New("want")
		loggedInUser = &uuid.User{
			Uuid: "fake-user-123",
		}

		alistUUID = "fake-list-123"
		userUUID = loggedInUser.Uuid
		dripfeedUUID = dripfeed.UUID(userUUID, alistUUID)

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
			uri                  = "/api/v1/spaced-repetition/overtime"
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
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "User doesnt match")
		})

		It("Failed to look up acl access", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
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

			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
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

			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
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

			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
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

			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
			aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
			aList := alist.NewTypeV3()
			aList.Uuid = alistUUID

			listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
			service.Create(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			// Hardcoding to catch when I change it
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Kind not supported: v1,v2")
		})

		It("Issue looking up to see if dripfeed exists", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

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

			dripfeedRepo.On("Exists", dripfeedUUID).Return(true, nil)
			service.Create(c)
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(dripfeedHTTPResponse))
		})

		When("List found but is empty", func() {
			BeforeEach(func() {
				b, _ := json.Marshal(input)
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
				c = e.NewContext(req, rec)

				c.Set("loggedInUser", *loggedInUser)
				c.SetPath(uri)

				dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
				aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
			})

			It("v1", func() {
				aList := alist.NewTypeV1()
				aList.Uuid = alistUUID

				listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
				service.Create(c)
				Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.SpacedRepetitionOvertimeEmptyList)
			})

			It("v2", func() {
				aList := alist.NewTypeV2()
				aList.Uuid = alistUUID

				listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)
				service.Create(c)
				Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.SpacedRepetitionOvertimeEmptyList)
			})

		})

		When("We will add over time", func() {
			It("v1", func() {
				b, _ := json.Marshal(input)
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
				c = e.NewContext(req, rec)

				c.Set("loggedInUser", *loggedInUser)
				c.SetPath(uri)

				dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
				aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
				aList := alist.NewTypeV1()
				aList.Uuid = input.AlistUuid
				aList.Data = append(aList.Data.(alist.TypeV1), "hello")

				listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)

				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiSpacedRepetitionOvertime))
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

					dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)

					aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
					aList := alist.NewTypeV2()
					aList.Uuid = input.AlistUuid
					aList.Data = append(aList.Data.(alist.TypeV2), alist.TypeV2Item{
						From: "car",
						To:   "bil",
					})

					listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)

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

					dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)
					aclRepo.On("HasUserListReadAccess", input.AlistUuid, input.UserUuid).Return(true, nil)
					aList := alist.NewTypeV2()
					aList.Uuid = input.AlistUuid
					aList.Data = append(aList.Data.(alist.TypeV2), alist.TypeV2Item{
						From: "car",
						To:   "bil",
					})

					listRepo.On("GetAlist", input.AlistUuid).Return(aList, nil)

					eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.ApiSpacedRepetitionOvertime))
						Expect(moment.Action).To(Equal(event.ActionCreated))

						b, _ := json.Marshal(moment.Data)
						var data dripfeed.EventDripfeedInputV2
						json.Unmarshal(b, &data)

						Expect(data.Info.AlistUUID).To(Equal(input.AlistUuid))
						Expect(data.Info.UserUUID).To(Equal(input.UserUuid))
						Expect(data.Info.Kind).To(Equal(alist.FromToList))
						return true
					}))

					service.Create(c)

					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(dripfeedHTTPResponse))
				})
			})
		})
	})

	When("Delete", func() {
		var (
			uri              = "/api/v1/spaced-repetition/overtime"
			inputFake, input openapi.SpacedRepetitionOvertimeInputBase
		)

		BeforeEach(func() {
			//dripfeedHTTPResponse = `{"dripfeed_uuid":"f29a45249551ae992a8edc6526ca7421094c8883","alist_uuid":"fake-list-123","user_uuid":"fake-user-123"}`
			inputFake = openapi.SpacedRepetitionOvertimeInputBase{
				AlistUuid: alistUUID,
				UserUuid:  "fake-user-456",
			}
			input = openapi.SpacedRepetitionOvertimeInputBase{
				AlistUuid: alistUUID,
				UserUuid:  userUUID,
			}
			fmt.Println(input)
		})

		It("User is not the one logged in", func() {
			b, _ := json.Marshal(inputFake)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)
			service.Delete(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "User doesnt match")
		})

		It("Issue looking up to see if dripfeed exists", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, want)

			service.Delete(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("No list to delete", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)

			service.Delete(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "List removed")
		})

		It("list queued for removal", func() {
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *loggedInUser)
			c.SetPath(uri)

			dripfeedRepo.On("Exists", dripfeedUUID).Return(true, nil)

			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(event.ApiSpacedRepetitionOvertime))
				Expect(moment.Action).To(Equal(event.ActionDeleted))
				Expect(moment.Data.(openapi.SpacedRepetitionOvertimeInfo).DripfeedUuid).To(Equal(dripfeedUUID))
				Expect(moment.Data.(openapi.SpacedRepetitionOvertimeInfo).UserUuid).To(Equal(userUUID))
				return true
			}))

			service.Delete(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "List removed")
		})
	})

	When("ListActive", func() {
		It("Issue looking up to see if dripfeed exists", func() {
			alistUUID := "fake-list-123"
			uri := fmt.Sprintf("/overtime/active/%s", alistUUID)
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/overtime/active/:alistUUID")
			c.Set("loggedInUser", *loggedInUser)
			c.SetParamNames("alistUUID")
			c.SetParamValues(alistUUID)

			dripfeedUUID := dripfeed.UUID(loggedInUser.Uuid, alistUUID)
			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, want)

			service.ListActive(c)

			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("List cant be found", func() {
			alistUUID := "fake-list-123"
			uri := fmt.Sprintf("/overtime/active/%s", alistUUID)
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/overtime/active/:alistUUID")
			c.Set("loggedInUser", *loggedInUser)
			c.SetParamNames("alistUUID")
			c.SetParamValues(alistUUID)

			dripfeedUUID := dripfeed.UUID(loggedInUser.Uuid, alistUUID)
			dripfeedRepo.On("Exists", dripfeedUUID).Return(false, nil)

			service.ListActive(c)

			Expect(rec.Code).To(Equal(http.StatusNotFound))
			Expect(rec.Body.Len()).To(Equal(0))
		})

		It("List found", func() {
			alistUUID := "fake-list-123"
			uri := fmt.Sprintf("/overtime/active/%s", alistUUID)
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/overtime/active/:alistUUID")
			c.Set("loggedInUser", *loggedInUser)
			c.SetParamNames("alistUUID")
			c.SetParamValues(alistUUID)

			dripfeedUUID := dripfeed.UUID(loggedInUser.Uuid, alistUUID)
			dripfeedRepo.On("Exists", dripfeedUUID).Return(true, nil)

			service.ListActive(c)

			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(rec.Body.Len()).To(Equal(0))
		})

	})
})
