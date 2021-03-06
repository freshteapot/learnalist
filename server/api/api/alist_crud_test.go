package api_test

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Alist endpoints", func() {
	AfterEach(emptyDatabase)

	When("Basic crud", func() {
		var userUUID string
		var datastore *mocks.Datastore

		var acl *mocks.Acl
		var user *uuid.User
		var method string
		var uri string
		BeforeEach(func() {
			method = http.MethodPost
			uri = "/api/v1/alist"

			datastore = &mocks.Datastore{}
			acl = &mocks.Acl{}
			m.Datastore = datastore
			m.Acl = acl

			userUUID = "fake-123"
			user = &uuid.User{
				Uuid: userUUID,
			}
		})
		Context("Save a list", func() {
			It("Reject if wrong method", func() {
				method = "DELETE"
				input := ""
				req, rec := setupFakeEndpoint(method, uri, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "This method is not supported.")
			})

			It("Get, is not accepted", func() {
				input := ""
				req, rec := setupFakeEndpoint(method, uri, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Your input is invalid json.")
			})

			It("Post, success", func() {
				savedList := alist.NewTypeV1()
				savedList.Uuid = "fake-list-123"
				datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(savedList, nil)
				datastore.On("GetAllListsByUser", user.Uuid).Return([]alist.ShortInfo{}, nil)
				datastore.On("GetPublicLists").Return([]alist.ShortInfo{}, nil)

				input := `
      {
      	"data": ["car"],
      	"info": {
      		"title": "Days of the Week",
      		"type": "v1"
      	}
      }
      `
				user := &uuid.User{
					Uuid: userUUID,
				}

				eventMessageBus := &mocks.EventlogPubSub{}
				eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
					Expect(moment.Kind).To(Equal(event.ApiListSaved))
					Expect(moment.Data.(event.EventList).UUID).To(Equal(savedList.Uuid))
					Expect(moment.Data.(event.EventList).Action).To(Equal("created"))
					return true
				}))
				event.SetBus(eventMessageBus)

				req, rec := setupFakeEndpoint(method, uri, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusCreated))
				b, _ := json.Marshal(savedList)

				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
			})

			It("Post, fail, due to ownership", func() {
				datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(alist.Alist{}, i18n.ErrorInputSaveAlistOperationOwnerOnly)
				input := `
      {
      	"data": ["car"],
      	"info": {
      		"title": "Days of the Week",
      		"type": "v1"
      	}
      }
      `
				user := &uuid.User{
					Uuid: userUUID,
				}

				req, rec := setupFakeEndpoint(method, uri, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusForbidden))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Only the owner of the list can modify it.")
			})

			It("PUT, fail, due to list uuid not being found", func() {
				method := http.MethodPut
				datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(nil, i18n.ErrorListNotFound)
				input := `
      {
      	"data": ["car"],
      	"info": {
      		"title": "Days of the Week",
      		"type": "v1"
      	}
      }
      `
				user := &uuid.User{
					Uuid: userUUID,
				}
				uri = uri + "/1234"
				req, rec := setupFakeEndpoint(method, "/", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.SetPath("/api/v1/alist/:uuid")
				c.Set("loggedInUser", *user)
				c.SetParamNames("uuid")
				c.SetParamValues("")
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Please refer to the documentation on lists")
			})

			It("PUT, fail, due to list uuid not being found", func() {
				method := http.MethodPut
				datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(alist.Alist{}, i18n.ErrorListNotFound)
				input := `
      {
      	"data": ["car"],
      	"info": {
      		"title": "Days of the Week",
      		"type": "v1"
      	}
      }
      `
				user := &uuid.User{
					Uuid: userUUID,
				}

				req, rec := setupFakeEndpoint(method, "/", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				c.SetPath("/alist/:uuid")
				c.Set("loggedInUser", *user)
				c.SetParamNames("uuid")
				c.SetParamValues("1234")
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusNotFound))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.SuccessAlistNotFound)
			})

			It("PUT, fail, due to uuid in uri not matching in the list", func() {
				method := http.MethodPut
				datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(nil, i18n.ErrorListNotFound)
				input := `
      {
      	"data": ["car"],
      	"info": {
      		"title": "Days of the Week",
      		"type": "v1"
      	},
				"uuid": "fake-456"
      }
      `
				user := &uuid.User{
					Uuid: userUUID,
				}

				req, rec := setupFakeEndpoint(method, "/", input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				c.SetPath("/alist/:uuid")
				c.Set("loggedInUser", *user)
				c.SetParamNames("uuid")
				c.SetParamValues("1234")
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "The list uuid in the uri doesnt match that in the payload")
			})

			It("Post, fail, due to internal issues", func() {
				datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(alist.Alist{}, errors.New("Failed"))
				input := `
      {
      	"data": ["car"],
      	"info": {
      		"title": "Days of the Week",
      		"type": "v1"
      	}
      }
      `
				user := &uuid.User{
					Uuid: userUUID,
				}

				req, rec := setupFakeEndpoint(method, uri, input)
				e := echo.New()
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1SaveAlist(c)
				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Failed")
			})
		})

	})

})
