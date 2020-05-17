package api_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"

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
	BeforeEach(func() {
		testHugoHelper := &mocks.HugoSiteBuilder{}
		testHugoHelper.On("Write", mock.Anything)
		testHugoHelper.On("Remove", mock.Anything)
		m.HugoHelper = testHugoHelper

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
			Expect(cleanEchoJSONResponse(rec)).To(Equal(`[]`))
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
			Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"The uuid is missing."}`))
		})

		It("Failed to lookup user access to the list, due to issues within", func() {
			aList := alist.NewTypeV1()
			aList.Uuid = alistUUID
			uri = fmt.Sprintf("/api/v1/alist/%s", alistUUID)
			req, rec := setupFakeEndpoint(method, uri, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/alist/:uuid")
			c.Set("loggedInUser", *user)
			c.SetParamNames("uuid")
			c.SetParamValues(alistUUID)

			acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(false, errors.New("Error"))
			m.V1GetListByUUID(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Issue with talking to the database whilst doing acl lookup"}`))
		})

		It("You do not have access to the list", func() {
			aList := alist.NewTypeV1()
			aList.Uuid = alistUUID
			uri = fmt.Sprintf("/api/v1/alist/%s", alistUUID)
			req, rec := setupFakeEndpoint(method, uri, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/alist/:uuid")
			c.Set("loggedInUser", *user)
			c.SetParamNames("uuid")
			c.SetParamValues(alistUUID)

			acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(false, nil)
			m.V1GetListByUUID(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Access Denied"}`))
		})

		// With the look up of access before, this one doesnt fully make sense anymore.
		It("Have access but an error something went wrong retrieving the list", func() {
			aList := alist.NewTypeV1()
			aList.Uuid = alistUUID
			uri = fmt.Sprintf("/api/v1/alist/%s", alistUUID)
			req, rec := setupFakeEndpoint(method, uri, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/alist/:uuid")
			c.Set("loggedInUser", *user)
			c.SetParamNames("uuid")
			c.SetParamValues(alistUUID)

			acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(true, nil)
			datastore.On("GetAlist", alistUUID).Return(aList, errors.New("Trigger"))
			m.V1GetListByUUID(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"message":"Failed to find alist with uuid: fake-list-123"}`))
		})

		It("Success", func() {
			aList := alist.NewTypeV1()
			aList.Uuid = alistUUID
			uri = fmt.Sprintf("/api/v1/alist/%s", alistUUID)
			req, rec := setupFakeEndpoint(method, uri, input)
			e := echo.New()
			c := e.NewContext(req, rec)
			c.SetPath("/api/v1/alist/:uuid")
			c.Set("loggedInUser", *user)
			c.SetParamNames("uuid")
			c.SetParamValues(alistUUID)

			acl.On("HasUserListReadAccess", alistUUID, user.Uuid).Return(true, nil)
			datastore.On("GetAlist", alistUUID).Return(aList, nil)
			m.V1GetListByUUID(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(cleanEchoJSONResponse(rec)).To(Equal(`{"data":[],"info":{"title":"","type":"v1","labels":[],"interact":{"slideshow":0,"totalrecall":0},"shared_with":"private"},"uuid":"fake-list-123"}`))

		})
	})
})
