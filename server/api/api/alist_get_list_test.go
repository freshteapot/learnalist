package api_test

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Api endpoints that get lists", func() {
	var (
		logger    *logrus.Logger
		hook      *test.Hook
		datastore *mocks.Datastore
		acl       *mocks.Acl
		user      *uuid.User
		method    string
		uri       string
		input     string
		alistUUID string
	)
	BeforeEach(func() {
		logger, hook = test.NewNullLogger()
		m.SetLogger(logger)
		datastore = &mocks.Datastore{}
		acl = &mocks.Acl{}
		m.Datastore = datastore
		m.Acl = acl

		alistUUID = "fake-list-123"
		method = http.MethodGet
		input = ""
		user = &uuid.User{
			Uuid: "fake-123",
		}
	})

	When("Get list by the user logged in", func() {
		It("No lists found", func() {

			var aLists = []alist.Alist{}
			uri = "/api/v1/alist/by/me"
			req, rec := setupFakeEndpoint(method, uri, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/alist/by/me")
			c.Set("loggedInUser", *user)
			datastore.On("GetListsByUserWithFilters", user.Uuid, "", "").Return(aLists, nil)

			m.V1GetListsByMe(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(cleanEchoResponse(rec)).To(Equal(`[]`))
		})
	})

	When("Get list by uuid", func() {

		It("No uuid", func() {
			alistUUID = ""
			uri = fmt.Sprintf("/api/v1/alist/%s", alistUUID)

			req, rec := setupFakeEndpoint(method, uri, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/alist/:uuid")
			c.Set("loggedInUser", *user)
			c.SetParamNames("uuid")
			c.SetParamValues(alistUUID)

			m.V1GetListByUUID(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			Expect(cleanEchoResponse(rec)).To(Equal(`{"message":"The uuid is missing."}`))
		})

		When("Request is valid", func() {
			var (
				c     echo.Context
				req   *http.Request
				rec   *httptest.ResponseRecorder
				aList alist.Alist
			)

			BeforeEach(func() {
				aList = alist.NewTypeV1()
				aList.Uuid = alistUUID
				uri = fmt.Sprintf("/api/v1/alist/%s", alistUUID)
				req, rec = setupFakeEndpoint(method, uri, input)
				e := echo.New()
				c = e.NewContext(req, rec)
				c.SetPath("/api/v1/alist/:uuid")
				c.Set("loggedInUser", *user)
				c.SetParamNames("uuid")
				c.SetParamValues(alistUUID)
			})

			It("Failed to lookup user access to the list, due to issues within", func() {
				acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(false, errors.New("Error"))
				m.V1GetListByUUID(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorAclLookup)
			})

			It("You do not have access to the list", func() {
				acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(false, nil)
				m.V1GetListByUUID(c)
				Expect(rec.Code).To(Equal(http.StatusForbidden))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.AclHttpAccessDeny)
			})

			// With the look up of access before, this one doesnt fully make sense anymore.
			It("Have access but an error something went wrong retrieving the list", func() {
				acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(true, nil)
				datastore.On("GetAlist", alistUUID).Return(alist.Alist{}, errors.New("Trigger"))
				m.V1GetListByUUID(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
			})

			It("Has access via acl, but list not found", func() {
				acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(true, nil)
				datastore.On("GetAlist", alistUUID).Return(alist.Alist{}, i18n.ErrorListNotFound)
				m.V1GetListByUUID(c)
				Expect(rec.Code).To(Equal(http.StatusNotFound))
				testutils.CheckMessageResponseFromResponseRecorder(rec, fmt.Sprintf(i18n.ApiAlistNotFound, alistUUID))
				Expect(hook.LastEntry().Data["event"]).To(Equal("broken-state"))
			})

			It("Success", func() {
				acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(true, nil)
				datastore.On("GetAlist", alistUUID).Return(aList, nil)
				m.V1GetListByUUID(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(cleanEchoResponse(rec)).To(Equal(`{"data":[],"info":{"title":"","type":"v1","labels":[],"interact":{"slideshow":0,"totalrecall":0},"shared_with":"private"},"uuid":"fake-list-123"}`))
			})
		})
	})
})
