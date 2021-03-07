package api_test

import (
	"encoding/json"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	aclKeys "github.com/freshteapot/learnalist-api/server/pkg/acl/keys"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing When a list is saved with a valid info.from", func() {
	AfterEach(emptyDatabase)
	BeforeEach(func() {
		eventMessageBus := &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		eventMessageBus.On("Publish", event.TopicMonolog, mock.Anything, mock.Anything)
	})

	var (
		datastore      *mocks.Datastore
		userManagement *mocks.Management
		acl            *mocks.Acl
		userA          *uuid.User
		userB          *uuid.User
		method         string
		uri            string
		e              *echo.Echo
		input          string
		inputGrant     api.HTTPShareListWithUserInput
		inputObject    api.HTTPShareListInput
	)

	BeforeEach(func() {
		datastore = &mocks.Datastore{}
		userManagement = &mocks.Management{}
		acl = &mocks.Acl{}
		m.Datastore = datastore
		m.UserManagement = userManagement
		m.Acl = acl

		userA = &uuid.User{
			Uuid: "fake-123",
		}
		userB = &uuid.User{
			Uuid: "fake-456",
		}

		inputGrant = api.HTTPShareListWithUserInput{
			UserUUID:  userB.Uuid,
			AlistUUID: "fakeList",
			Action:    aclKeys.ActionGrant,
		}
		a, _ := json.Marshal(inputGrant)
		input = string(a)
	})

	When("/share/readaccess", func() {
		BeforeEach(func() {
			method = http.MethodPost
			uri = "/api/v1/share/readaccess"
			e = echo.New()
		})

		Context("Kind is not learnalist", func() {
			It("Reject if shared is not private", func() {
				req, rec := setupFakeEndpoint(method, uri, input)
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *userA)

				aList := alist.NewTypeV1()
				aList.User.Uuid = userA.Uuid
				aList.Info.SharedWith = aclKeys.SharedWithPublic
				aList.Info.From = &openapi.AlistFrom{
					Kind:    "cram",
					RefUrl:  "https://cram.com/xxx",
					ExtUuid: "xxx",
				}
				datastore.On("GetAlist", mock.Anything).Return(aList, nil)
				datastore.On("UserExists", userB.Uuid).Return(true)
				acl.On("GrantUserListReadAccess", inputGrant.AlistUUID, inputGrant.UserUUID).Return(nil)

				m.V1ShareListReadAccess(c)

				Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InputSaveAlistOperationFromRestriction)
			})
		})

		It("Allow learnalist", func() {
			req, rec := setupFakeEndpoint(method, uri, input)
			c := e.NewContext(req, rec)
			c.Set("loggedInUser", *userA)

			aList := alist.NewTypeV1()
			aList.User.Uuid = userA.Uuid
			aList.Info.SharedWith = aclKeys.SharedWithPublic
			aList.Info.From = &openapi.AlistFrom{
				Kind:    "learnalist",
				RefUrl:  "https://learnalist.net/xxx",
				ExtUuid: "xxx",
			}
			datastore.On("GetAlist", mock.Anything).Return(aList, nil)
			userManagement.On("UserExists", userB.Uuid).Return(true)
			acl.On("GrantUserListReadAccess", inputGrant.AlistUUID, inputGrant.UserUUID).Return(nil)

			m.V1ShareListReadAccess(c)

			Expect(rec.Code).To(Equal(http.StatusOK))
			response := testutils.CleanEchoResponseFromResponseRecorder(rec)
			Expect(response).To(Equal(input))
		})
	})

	When("/share", func() {
		BeforeEach(func() {
			method = http.MethodPost
			uri = "/api/v1/share/alist"
			e = echo.New()

			inputObject = api.HTTPShareListInput{
				AlistUUID: "fakeList",
				Action:    aclKeys.SharedWithPublic,
			}
			a, _ := json.Marshal(inputObject)
			input = string(a)
		})

		Context("Kind is not learnalist", func() {
			It("Reject if shared is not private", func() {
				req, rec := setupFakeEndpoint(method, uri, input)
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *userA)

				aList := alist.NewTypeV1()
				aList.Uuid = inputObject.AlistUUID
				aList.User.Uuid = userA.Uuid
				aList.Info.From = &openapi.AlistFrom{
					Kind:    "cram",
					RefUrl:  "https://cram.com/xxx",
					ExtUuid: "xxx",
				}
				aList.Info.SharedWith = aclKeys.NotShared

				datastore.On("GetAlist", mock.Anything).Return(aList, nil)

				datastore.On("GetAllListsByUser", userA.Uuid).Return([]alist.ShortInfo{}, nil)
				datastore.On("GetPublicLists").Return([]alist.ShortInfo{})

				m.V1ShareAlist(c)

				Expect(rec.Code).To(Equal(http.StatusForbidden))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InputSaveAlistOperationFromRestriction)
			})
		})

		It("Allow learnalist", func() {
			req, rec := setupFakeEndpoint(method, uri, input)
			c := e.NewContext(req, rec)
			c.Set("loggedInUser", *userA)

			listInfoFrom := &openapi.AlistFrom{
				Kind:    "learnalist",
				RefUrl:  "https://learnalist.net/xxx",
				ExtUuid: "xxx",
			}

			aList := alist.NewTypeV1()
			aList.Uuid = inputObject.AlistUUID
			aList.User.Uuid = userA.Uuid

			aList.Info.From = listInfoFrom
			aList.Info.SharedWith = aclKeys.NotShared

			returnAlist := alist.NewTypeV1()
			returnAlist.Uuid = inputObject.AlistUUID
			returnAlist.User.Uuid = userA.Uuid
			returnAlist.Info.From = listInfoFrom
			returnAlist.Info.SharedWith = aclKeys.SharedWithPublic

			datastore.On("GetAlist", mock.Anything).Return(aList, nil)
			datastore.On("SaveAlist", http.MethodPut, returnAlist).Return(returnAlist, nil)
			datastore.On("GetAllListsByUser", userA.Uuid).Return([]alist.ShortInfo{}, nil)
			datastore.On("GetPublicLists").Return([]alist.ShortInfo{})

			m.V1ShareAlist(c)

			Expect(rec.Code).To(Equal(http.StatusOK))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.ApiShareListSuccessWithPublic)
		})
	})

	When("Updating a list in the system", func() {
		BeforeEach(func() {
			method = http.MethodPut
			uri = "/api/v1/alist/1234"
			e = echo.New()

			inputObject = api.HTTPShareListInput{
				AlistUUID: "fakeList",
				Action:    aclKeys.SharedWithPublic,
			}
			a, _ := json.Marshal(inputObject)
			input = string(a)
		})

		// This does not test the meat
		It("Kind not learnalist", func() {
			aList := alist.NewTypeV1()
			aList.Uuid = inputObject.AlistUUID
			aList.User.Uuid = userA.Uuid
			aList.Info.From = &openapi.AlistFrom{
				Kind:    "cram",
				RefUrl:  "https://cram.com/xxx",
				ExtUuid: "xxx",
			}
			aList.Info.SharedWith = aclKeys.SharedWithPublic

			b, _ := json.Marshal(aList)
			input = string(b)
			// Set back so we can use this object
			returnAlist := aList
			returnAlist.Info.SharedWith = aclKeys.NotShared
			req, rec := setupFakeEndpoint(method, uri, input)
			c := e.NewContext(req, rec)
			c.SetPath("/alist/:uuid")
			c.SetParamNames("uuid")
			c.SetParamValues(inputObject.AlistUUID)
			c.Set("loggedInUser", *userA)

			datastore.On("GetAlist", mock.Anything).Return(returnAlist, nil)
			// TODO maybe we need some errors
			datastore.On("SaveAlist", http.MethodPut, aList).Return(aList, i18n.ErrorInputSaveAlistOperationFromRestriction)
			m.V1SaveAlist(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InputSaveAlistOperationFromRestriction)
		})

		// TODO when is learnalist
		// TODO do we tidy the whole of alist_crud_test.go?
	})
})
