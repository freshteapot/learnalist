package api_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"

	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Label endpoints", func() {
	AfterEach(emptyDatabase)

	When("/labels", func() {
		Context("Posting", func() {
			var (
				datastore    *mocks.Datastore
				acl          *mocks.Acl
				labelStorage *mocks.LabelReadWriter
				user         *uuid.User
				method       string
				uri          string
				e            *echo.Echo
			)

			BeforeEach(func() {
				labelStorage = &mocks.LabelReadWriter{}
				datastore = &mocks.Datastore{}
				datastore.On("Labels").Return(labelStorage)
				acl = &mocks.Acl{}
				m.Datastore = datastore
				m.Acl = acl

				user = &uuid.User{
					Uuid: "fake-123",
				}

				method = http.MethodPost
				uri = "/api/v1/labels"
				e = echo.New()
			})

			It("Invalid json input", func() {
				input := ""
				req, rec := setupFakeEndpoint(method, uri, input)
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				m.V1PostUserLabel(c)

				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Your input is invalid json.")
			})

			It("Valid json, invalid input", func() {
				input := api.HTTPLabelInput{
					Label: "",
				}
				b, _ := json.Marshal(input)

				req, rec := setupFakeEndpoint(method, uri, string(b))
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)

				labelStorage.On("PostUserLabel", mock.Anything).Return(http.StatusBadRequest, errors.New("Fail"))
				m.V1PostUserLabel(c)

				Expect(rec.Code).To(Equal(http.StatusBadRequest))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Please refer to the documentation on label(s)")
			})

			It("Valid input, failed to save", func() {
				input := api.HTTPLabelInput{
					Label: "I am a label",
				}
				b, _ := json.Marshal(input)

				req, rec := setupFakeEndpoint(method, uri, string(b))
				c := e.NewContext(req, rec)
				c.Set("loggedInUser", *user)

				labelStorage.On("PostUserLabel", mock.Anything).Return(http.StatusInternalServerError, errors.New("Fail"))
				m.V1PostUserLabel(c)

				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Sadly, our service has taken a nap.")
			})

			Context("Success, label saved", func() {
				It("Its new", func() {
					input := api.HTTPLabelInput{
						Label: "I am a label",
					}
					b, _ := json.Marshal(input)

					req, rec := setupFakeEndpoint(method, uri, string(b))
					c := e.NewContext(req, rec)
					c.Set("loggedInUser", *user)

					labelStorage.On("PostUserLabel", mock.Anything).Return(http.StatusCreated, nil)
					labelStorage.On("GetUserLabels", mock.Anything).Return([]string{input.Label}, nil)
					m.V1PostUserLabel(c)

					Expect(rec.Code).To(Equal(http.StatusCreated))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`["I am a label"]`))
				})

				It("Its already in the system", func() {
					input := api.HTTPLabelInput{
						Label: "I am a label",
					}
					b, _ := json.Marshal(input)

					req, rec := setupFakeEndpoint(method, uri, string(b))
					c := e.NewContext(req, rec)
					c.Set("loggedInUser", *user)

					labelStorage.On("PostUserLabel", mock.Anything).Return(http.StatusOK, nil)
					labelStorage.On("GetUserLabels", mock.Anything).Return([]string{input.Label}, nil)
					m.V1PostUserLabel(c)

					Expect(rec.Code).To(Equal(http.StatusOK))
					Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`["I am a label"]`))
				})
			})
		})
		Context("Get", func() {
			var (
				datastore    *mocks.Datastore
				labelStorage *mocks.LabelReadWriter
				acl          *mocks.Acl
				user         *uuid.User
				method       string
				uri          string
				req          *http.Request
				rec          *httptest.ResponseRecorder
				e            *echo.Echo
				c            echo.Context
			)

			BeforeEach(func() {
				labelStorage = &mocks.LabelReadWriter{}
				datastore = &mocks.Datastore{}
				datastore.On("Labels").Return(labelStorage)
				acl = &mocks.Acl{}
				m.Datastore = datastore
				m.Acl = acl

				user = &uuid.User{
					Uuid: "fake-123",
				}

				method = http.MethodGet
				uri = "/api/v1/labels/by/me"
				e = echo.New()
				req, rec = setupFakeEndpoint(method, uri, "")
				c = e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
			})

			It("User with no labels", func() {
				labelStorage.On("GetUserLabels", mock.Anything).Return([]string{}, nil)

				m.V1GetUserLabels(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`[]`))
			})

			It("User with labels", func() {
				labelStorage.On("GetUserLabels", mock.Anything).Return([]string{
					"wind",
					"water",
					"fire",
					"earth",
				}, nil)

				m.V1GetUserLabels(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`["wind","water","fire","earth"]`))
			})

			It("Something went wrong getting the data from the storage", func() {
				labelStorage.On("GetUserLabels", mock.Anything).Return([]string{}, errors.New("Failed"))
				m.V1GetUserLabels(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Sadly, our service has taken a nap.")
			})
		})

		Context("DELETE", func() {
			var (
				datastore    *mocks.Datastore
				labelStorage *mocks.LabelReadWriter
				acl          *mocks.Acl
				user         *uuid.User
				method       string
				uri          string
				req          *http.Request
				rec          *httptest.ResponseRecorder
				e            *echo.Echo
				c            echo.Context
			)

			BeforeEach(func() {
				labelStorage = &mocks.LabelReadWriter{}
				datastore = &mocks.Datastore{}
				datastore.On("Labels").Return(labelStorage)
				acl = &mocks.Acl{}
				m.Datastore = datastore
				m.Acl = acl

				user = &uuid.User{
					Uuid: "fake-123",
				}

				method = http.MethodGet
			})

			It("Failed to remove a label, due to storage failing", func() {
				uri = "/api/v1/labels/test"
				e = echo.New()
				req, rec = setupFakeEndpoint(method, uri, "")
				c = e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				c.SetPath("/api/v1/labels/:label")
				c.Set("loggedInUser", *user)
				c.SetParamNames("label")
				c.SetParamValues("test")

				datastore.On("RemoveUserLabel", "test", user.Uuid).Return(errors.New("Failed"))
				m.V1RemoveUserLabel(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Sadly, our service has taken a nap.")
			})

			It("Successfully removed a label", func() {
				uri = "/api/v1/labels/test"
				e = echo.New()
				req, rec = setupFakeEndpoint(method, uri, "")
				c = e.NewContext(req, rec)
				c.Set("loggedInUser", *user)
				c.SetPath("/api/v1/labels/:label")
				c.Set("loggedInUser", *user)
				c.SetParamNames("label")
				c.SetParamValues("test")

				datastore.On("RemoveUserLabel", "test", user.Uuid).Return(nil)
				m.V1RemoveUserLabel(c)
				Expect(rec.Code).To(Equal(http.StatusOK))
				testutils.CheckMessageResponseFromResponseRecorder(rec, "Label test was removed.")
			})
		})
	})
})
