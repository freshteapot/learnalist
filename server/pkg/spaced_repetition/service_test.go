package spaced_repetition_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/alist"
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
	When("Saving an entry", func() {
		It("Not valid entry", func() {
			uri := "/api/v1//api/v1/spaced-repetition"
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
			uri := "/api/v1//api/v1/spaced-repetition"
			input := `
			{
				"show": "Hello",
				"data": "Hello",
				"kind": "v1"
			  }
			`
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			repo.On("SaveEntry", mock.Anything).Return(spaced_repetition.ErrSpacedRepetitionEntryExists)
			expect := openapi.SpacedRepetitionV1{
				Kind: alist.SimpleList,
				Show: "Hello",
				Data: "Hello",
			}

			repo.On("GetEntry", user.Uuid, "ba9277fc4c6190fb875ad8f9cee848dba699937f").Return(expect, nil)

			service.SaveEntry(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			b, _ := json.Marshal(expect)
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
		})

		It("New Entry", func() {
			uri := "/api/v1//api/v1/spaced-repetition"
			input := `
			{
				"show": "Hello",
				"data": "Hello",
				"kind": "v1"
			  }
			`
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			repo.On("SaveEntry", mock.Anything).Return(nil)
			expect := openapi.SpacedRepetitionV1{
				Kind: alist.SimpleList,
				Show: "Hello",
				Data: "Hello",
			}
			entryUUID := "ba9277fc4c6190fb875ad8f9cee848dba699937f"
			userUUID := user.Uuid
			repo.On("GetEntry", userUUID, entryUUID).Return(expect, nil)
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
			b, _ := json.Marshal(expect)
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
		})
	})
})
