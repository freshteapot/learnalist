package plank_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/i18n"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/mocks"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
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
		service plank.PlankService
		repo    *mocks.PlankRepository
		logger  *logrus.Logger
		c       echo.Context
		req     *http.Request
		rec     *httptest.ResponseRecorder
		user    *uuid.User
	)

	When("Requesting history", func() {
		BeforeEach(func() {
			user = &uuid.User{
				Uuid: "fake-123",
			}

			eventMessageBus := &mocks.EventlogPubSub{}
			eventMessageBus.On("Subscribe", mock.Anything, mock.Anything)
			event.SetBus(eventMessageBus)

			method := http.MethodPost
			uri := "/api/v1/plank/history"
			req, rec = setupFakeEndpoint(method, uri, "")
			e := echo.New()
			c = e.NewContext(req, rec)
			c.SetPath("/api/v1/plank/history")
			c.Set("loggedInUser", *user)

			logger, _ = test.NewNullLogger()
			repo = &mocks.PlankRepository{}
			service = plank.NewService(repo, logger)
		})

		It("Repo lookup failed", func() {
			want := errors.New("Fail")
			repo.On("History", user.Uuid).Return(make([]plank.HttpRequestInput, 0), want)
			service.History(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("No history", func() {
			repo.On("History", user.Uuid).Return(make([]plank.HttpRequestInput, 0), nil)
			service.History(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			Expect(testutils.CleanEchoResponseFromResponseRecorder(rec)).To(Equal(`[]`))
		})

		It("One entry", func() {
			rawRecords := `[{"beginningTime":1602264153548,"currentTime":1602264219291,"intervalTime":15,"intervalTimerNow":5681,"laps":4,"showIntervals":true,"timerNow":65743,"uuid":"98952178a3d23a356a396ffdc9e726629f80b8a8"}]`
			var expect []plank.HttpRequestInput
			var records []plank.HttpRequestInput

			json.Unmarshal([]byte(rawRecords), &records)
			repo.On("History", user.Uuid).Return(records, nil)
			service.History(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			json.Unmarshal(rec.Body.Bytes(), &expect)
			Expect(records).To(Equal(expect))
		})
	})
})

func setupFakeEndpoint(method string, uri string, body string) (*http.Request, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, uri, strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	return req, rec
}
