package spaced_repetition_test

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
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Spaced Repetition Service API", func() {
	var (
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder
		service         spaced_repetition.SpacedRepetitionService
		repo            *mocks.SpacedRepetitionRepository
		user            *uuid.User
	)

	BeforeEach(func() {
		user = &uuid.User{
			Uuid: "fake-123",
		}
		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		e = echo.New()

		logger, _ = test.NewNullLogger()
		repo = &mocks.SpacedRepetitionRepository{}
		service = spaced_repetition.NewService(repo, logger)
	})

	When("Deleting an entry", func() {
		var (
			uri   = "/api/v1/api/v1/spaced-repetition/ba9277fc4c6190fb875ad8f9cee848dba699937f"
			input = `
			{
				"show": "Hello",
				"data": "Hello",
				"kind": "v1"
			  }
			`
			entryUUID = "ba9277fc4c6190fb875ad8f9cee848dba699937f"
		)

		It("Missing the entry id", func() {
			fmt.Printf(entryUUID)
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, input)
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath("/api/v1/api/v1/spaced-repetition/:uuid")
			c.SetParamNames("uuid")
			c.SetParamValues("")

			service.DeleteEntry(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InputMissingListUuid)
		})

		It("Entry not found", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, input)
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath("/api/v1/api/v1/spaced-repetition/:uuid")
			c.SetParamNames("uuid")
			c.SetParamValues(entryUUID)

			repo.On("GetEntry", user.Uuid, entryUUID).Return(nil, spaced_repetition.ErrNotFound)
			service.DeleteEntry(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
		})

		It("Entry failed due to repo lookup", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, input)
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath("/api/v1/api/v1/spaced-repetition/:uuid")
			c.SetParamNames("uuid")
			c.SetParamValues(entryUUID)

			repo.On("GetEntry", user.Uuid, entryUUID).Return(nil, errors.New("fail"))
			service.DeleteEntry(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})
	})
	When("Saving an entry", func() {
		var (
			uri   = "/api/v1//api/v1/spaced-repetition"
			input = `
			{
				"show": "Hello",
				"data": "Hello",
				"kind": "v1"
			  }
			`
			entryUUID = "ba9277fc4c6190fb875ad8f9cee848dba699937f"
		)
		It("Not valid entry", func() {
			input := `
			{
				"show": "",
				"data": "",
				"kind": "v3"
			  }
			`
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)
			service.SaveEntry(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Kind not supported: v1,v2")
		})

		It("Entry already exists", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			repo.On("SaveEntry", mock.Anything).Return(spaced_repetition.ErrSpacedRepetitionEntryExists)
			service.SaveEntry(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var entry openapi.SpacedRepetitionV1
			json.Unmarshal(rec.Body.Bytes(), &entry)

			Expect(entry.Uuid).To(Equal(entryUUID))
			Expect(entry.Kind).To(Equal(alist.SimpleList))
			Expect(entry.Show).To(Equal("Hello"))
			Expect(entry.Settings.Level).To(Equal("0"))
		})

		It("New Entry", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			repo.On("SaveEntry", mock.Anything).Return(nil)
			userUUID := user.Uuid

			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(event.ApiSpacedRepetition))
				b, _ := json.Marshal(moment.Data)
				var data spaced_repetition.EventSpacedRepetition
				json.Unmarshal(b, &data)

				Expect(data.Kind).To(Equal(spaced_repetition.EventKindNew))
				Expect(data.Data.UserUUID).To(Equal(userUUID))
				Expect(data.Data.UUID).To(Equal(entryUUID))
				return true
			}))

			service.SaveEntry(c)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			var entry openapi.SpacedRepetitionV1
			json.Unmarshal(rec.Body.Bytes(), &entry)

			Expect(entry.Uuid).To(Equal(entryUUID))
			Expect(entry.Kind).To(Equal(alist.SimpleList))
			Expect(entry.Show).To(Equal("Hello"))
			Expect(entry.Settings.Level).To(Equal("0"))
		})

		It("Failed to save entry", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			want := errors.New("fail")
			repo.On("SaveEntry", mock.Anything).Return(want)
			service.SaveEntry(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
