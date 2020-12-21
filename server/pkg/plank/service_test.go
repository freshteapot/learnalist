package plank_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/plank"
	"github.com/freshteapot/learnalist-api/server/pkg/testutils"
	"github.com/labstack/echo/v4"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing API", func() {
	var (
		sampleRecords   = `[{"beginningTime":1602264153548,"currentTime":1602264219291,"intervalTime":15,"intervalTimerNow":5681,"laps":4,"showIntervals":true,"timerNow":65743,"uuid":"98952178a3d23a356a396ffdc9e726629f80b8a8"}]`
		service         plank.PlankService
		repo            *mocks.PlankRepository
		eventMessageBus *mocks.EventlogPubSub
		logger          *logrus.Logger
		c               echo.Context
		e               *echo.Echo
		req             *http.Request
		rec             *httptest.ResponseRecorder
		user            *uuid.User
		records         []openapi.Plank
		record          openapi.Plank
	)

	BeforeEach(func() {
		user = &uuid.User{
			Uuid: "fake-123",
		}
		eventMessageBus = &mocks.EventlogPubSub{}
		eventMessageBus.On("Subscribe", event.TopicMonolog, "plank", mock.Anything)
		event.SetBus(eventMessageBus)

		e = echo.New()

		logger, _ = test.NewNullLogger()
		repo = &mocks.PlankRepository{}
		service = plank.NewService(repo, logger)

		json.Unmarshal([]byte(sampleRecords), &records)
		record = records[0]
	})

	When("Requesting history", func() {
		BeforeEach(func() {
			method := http.MethodGet
			uri := "/api/v1/plank/history"
			req, rec = testutils.SetupJSONEndpoint(method, uri, "")
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/plank/history")
			c.Set("loggedInUser", *user)
		})

		It("Repo lookup failed", func() {
			want := errors.New("Fail")
			repo.On("History", user.Uuid).Return(make([]openapi.Plank, 0), want)
			service.History(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("No history", func() {
			repo.On("History", user.Uuid).Return(make([]openapi.Plank, 0), nil)
			service.History(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`[]`))
		})

		It("One entry", func() {
			var expect []openapi.Plank
			repo.On("History", user.Uuid).Return(records, nil)
			service.History(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			json.Unmarshal(rec.Body.Bytes(), &expect)
			Expect(records).To(Equal(expect))
		})
	})

	When("Adding a record", func() {
		BeforeEach(func() {
			b, _ := json.Marshal(record)
			method := http.MethodPost
			uri := "/api/v1/plank/"
			req, rec = testutils.SetupJSONEndpoint(method, uri, string(b))
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)
		})

		It("Failed to save", func() {
			want := errors.New("want")
			repo.On("SaveEntry", mock.Anything).Return(want)
			service.RecordPlank(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("saved", func() {
			repo.On("SaveEntry", mock.Anything).Return(nil)
			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Data.(event.EventPlank).UserUUID).To(Equal(user.Uuid))
				Expect(moment.Data.(event.EventPlank).Data).To(Equal(record))
				Expect(moment.Data.(event.EventPlank).Action).To(Equal(event.ActionNew))
				return true
			}))

			service.RecordPlank(c)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			b, _ := json.Marshal(record)
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
		})

		It("already recorded", func() {
			repo.On("SaveEntry", mock.Anything).Return(plank.ErrEntryExists)
			service.RecordPlank(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			b, _ := json.Marshal(record)
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(string(b)))
		})
	})

	When("Deleting a record", func() {
		var (
			recordUUID = "98952178a3d23a356a396ffdc9e726629f80b8a8"
			method     = http.MethodDelete
		)
		BeforeEach(func() {
			b, _ := json.Marshal(record)
			uri := fmt.Sprintf("/api/v1/plank/%s", recordUUID)
			req, rec = testutils.SetupJSONEndpoint(method, uri, string(b))
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/plank/:uuid")
			c.Set("loggedInUser", *user)
			c.SetParamNames("uuid")
			c.SetParamValues(recordUUID)
		})

		It("UUID is missing", func() {
			// Not sure this is possible
			c.SetParamNames("uuid")
			c.SetParamValues("")
			service.DeletePlankRecord(c)
			Expect(rec.Code).To(Equal(http.StatusBadRequest))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InputMissingListUuid)
		})

		It("Issue saving to repo", func() {
			want := errors.New("want")
			repo.On("GetEntry", recordUUID, user.Uuid).Return(openapi.Plank{}, want)
			service.DeletePlankRecord(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Record not found", func() {
			repo.On("GetEntry", recordUUID, user.Uuid).Return(openapi.Plank{}, plank.ErrNotFound)
			service.DeletePlankRecord(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.PlankRecordNotFound)
		})

		It("Record found, but failed to delete", func() {
			want := errors.New("want")
			repo.On("GetEntry", recordUUID, user.Uuid).Return(openapi.Plank{Uuid: recordUUID}, nil)
			repo.On("DeleteEntry", recordUUID, user.Uuid).Return(want)
			service.DeletePlankRecord(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("record removed", func() {
			repo.On("GetEntry", recordUUID, user.Uuid).Return(record, nil)
			repo.On("DeleteEntry", recordUUID, user.Uuid).Return(nil)

			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Data.(event.EventPlank).UserUUID).To(Equal(user.Uuid))
				Expect(moment.Data.(event.EventPlank).Data).To(Equal(record))
				Expect(moment.Data.(event.EventPlank).Action).To(Equal(event.ActionDeleted))
				return true
			}))

			service.DeletePlankRecord(c)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})
	})
})
