package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"

	"github.com/freshteapot/learnalist-api/server/api/uuid"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Api endpoints that get lists", func() {
	AfterEach(emptyDatabase)

	var datastore *mocks.Datastore
	var acl *mocks.Acl
	var user *uuid.User
	var (
		method    string
		uri       string
		input     string
		alistUUID string
	)

	var c echo.Context
	var req *http.Request
	var rec *httptest.ResponseRecorder
	BeforeEach(func() {

		datastore = &mocks.Datastore{}
		acl = &mocks.Acl{}
		m.Datastore = datastore
		m.Acl = acl

		alistUUID = "fake-list-123"
		method = http.MethodDelete
		input = ""
		user = &uuid.User{
			Uuid: "fake-123",
		}

		uri = fmt.Sprintf("/api/v1/alist/%s", alistUUID)
		req, rec = setupFakeEndpoint(method, uri, input)
		e := echo.New()
		c = e.NewContext(req, rec)
		c.SetPath("/api/v1/alist/:uuid")
		c.Set("loggedInUser", *user)
		c.SetParamNames("uuid")
		c.SetParamValues(alistUUID)
	})

	When("Remove a list", func() {
		It("List being removed is not found", func() {
			datastore.On("RemoveAlist", alistUUID, user.Uuid).Return(i18n.ErrorListNotFound)
			m.V1RemoveAlist(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "List not found.")
		})

		It("Only the owner of the list can remove it", func() {
			datastore.On("RemoveAlist", alistUUID, user.Uuid).Return(errors.New(i18n.InputDeleteAlistOperationOwnerOnly))
			m.V1RemoveAlist(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Only the owner of the list can remove it.")
		})

		It("An error occurred whilst trying to remove the list", func() {
			datastore.On("RemoveAlist", alistUUID, user.Uuid).Return(errors.New("Fail"))
			m.V1RemoveAlist(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "We have failed to remove your list.")
		})

		It("Successfully removed a list", func() {
			datastore.On("RemoveAlist", alistUUID, user.Uuid).Return(nil)
			datastore.On("GetAllListsByUser", user.Uuid).Return([]alist.ShortInfo{}, nil)
			datastore.On("GetPublicLists").Return([]alist.ShortInfo{}, nil)
			eventMessageBus := &mocks.EventlogPubSub{}
			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				//fmt.Println("moment.Kind", moment.Kind)
				//return true
				Expect(moment.Kind).To(Equal(event.ApiListDelete))
				Expect(moment.Data.(event.EventListOwner).UUID).To(Equal(alistUUID))
				return true
			}))
			event.SetBus(eventMessageBus)

			m.V1RemoveAlist(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "List fake-list-123 was removed.")
		})
	})
})
