package spaced_repetition_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"time"

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
		eventMessageBus                  *mocks.EventlogPubSub
		logger                           *logrus.Logger
		c                                echo.Context
		e                                *echo.Echo
		req                              *http.Request
		rec                              *httptest.ResponseRecorder
		service                          spaced_repetition.SpacedRepetitionService
		repo                             *mocks.SpacedRepetitionRepository
		user                             *uuid.User
		want                             error
		userUUID, entryUUID, entryUUIDV2 string
		inputV2                          = `
{
  "show": "Mars",
  "data": {
    "from": "March",
    "to": "Mars"
  },
  "settings": {
    "show": "to"
  },
  "kind": "v2"
}
`
	)

	BeforeEach(func() {
		entryUUIDV2 = "75698c0f5a7b904f1799ceb68e2afe67ad987689"
		entryUUID = "ba9277fc4c6190fb875ad8f9cee848dba699937f"
		want = errors.New("fail")
		user = &uuid.User{
			Uuid: "fake-123",
		}
		userUUID = user.Uuid

		eventMessageBus = &mocks.EventlogPubSub{}
		event.SetBus(eventMessageBus)
		e = echo.New()

		logger, _ = test.NewNullLogger()
		repo = &mocks.SpacedRepetitionRepository{}
		service = spaced_repetition.NewService(repo, logger)
	})

	When("Updating entry", func() {
		var (
			uri = "/api/v1/api/v1/spaced-repetition/viewed"
		)
		It("Check for invalid action", func() {
			input := openapi.SpacedRepetitionEntryViewed{
				Action: "fake",
				Uuid:   entryUUID,
			}
			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)
			service.EntryViewed(c)
			Expect(rec.Code).To(Equal(http.StatusUnprocessableEntity))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Action not supported: incr,decr")
		})

		It("Not found", func() {
			input := openapi.SpacedRepetitionEntryViewed{
				Action: spaced_repetition.ActionIncrement,
				Uuid:   entryUUID,
			}

			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)
			repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{}, spaced_repetition.ErrNotFound)

			service.EntryViewed(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
			Expect(len(rec.Body.Bytes())).To(Equal(0))
		})

		It("Error talking to repo to get the next entry", func() {
			input := openapi.SpacedRepetitionEntryViewed{
				Action: spaced_repetition.ActionIncrement,
				Uuid:   entryUUID,
			}

			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)
			repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{}, want)
			service.EntryViewed(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Entry found, but it is not the one currently being modified", func() {
			input := openapi.SpacedRepetitionEntryViewed{
				Action: spaced_repetition.ActionIncrement,
				Uuid:   entryUUID,
			}

			b, _ := json.Marshal(input)
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)
			repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{
				UUID: "something-else",
			}, nil)

			service.EntryViewed(c)
			Expect(rec.Code).To(Equal(http.StatusForbidden))
			testutils.CheckMessageResponseFromResponseRecorder(rec, "Input uuid is not the uuid of what is next")
		})

		When("updating entry based on action", func() {
			It("Failed to update via the repo", func() {
				input := openapi.SpacedRepetitionEntryViewed{
					Action: spaced_repetition.ActionIncrement,
					Uuid:   entryUUID,
				}

				b, _ := json.Marshal(input)
				req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
				c = e.NewContext(req, rec)

				c.Set("loggedInUser", *user)
				c.SetPath(uri)
				created, _ := time.Parse(time.RFC3339, "2020-12-27T17:04:59Z")
				whenNext, _ := time.Parse(time.RFC3339, "2020-12-27T18:04:59Z")
				repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{
					UUID:     entryUUID,
					UserUUID: userUUID,
					Body:     `{"show":"Hello","kind":"v1","uuid":"ba9277fc4c6190fb875ad8f9cee848dba699937f","data":"Hello","settings":{"level":"0","when_next":"2020-12-27T18:04:59Z","created":"2020-12-27T17:04:59Z"}}`,
					Created:  created,
					WhenNext: whenNext,
				}, nil)

				repo.On("UpdateEntry", mock.Anything).Return(want)

				service.EntryViewed(c)
				Expect(rec.Code).To(Equal(http.StatusInternalServerError))
				testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
			})

			When("Successful path", func() {
				It("Increment", func() {
					input := openapi.SpacedRepetitionEntryViewed{
						Action: spaced_repetition.ActionIncrement,
						Uuid:   entryUUID,
					}

					b, _ := json.Marshal(input)
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
					c = e.NewContext(req, rec)

					c.Set("loggedInUser", *user)
					c.SetPath(uri)
					created, _ := time.Parse(time.RFC3339, "2020-12-27T17:04:59Z")
					whenNext, _ := time.Parse(time.RFC3339, "2020-12-27T18:04:59Z")
					repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{
						UUID:     entryUUID,
						UserUUID: userUUID,
						Body:     `{"show":"Hello","kind":"v1","uuid":"ba9277fc4c6190fb875ad8f9cee848dba699937f","data":"Hello","settings":{"level":"0","when_next":"2020-12-27T18:04:59Z","created":"2020-12-27T17:04:59Z"}}`,
						Created:  created,
						WhenNext: whenNext,
					}, nil)

					repo.On("UpdateEntry", mock.Anything).Return(nil)
					eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.ApiSpacedRepetition))

						b, _ := json.Marshal(moment.Data)
						var data spaced_repetition.EventSpacedRepetition
						json.Unmarshal(b, &data)

						Expect(data.Kind).To(Equal(spaced_repetition.EventKindViewed))
						Expect(data.Data.UserUUID).To(Equal(userUUID))
						Expect(data.Data.UUID).To(Equal(entryUUID))

						return true
					}))

					service.EntryViewed(c)
					Expect(rec.Code).To(Equal(http.StatusOK))
					var entry openapi.SpacedRepetitionV1
					json.Unmarshal(rec.Body.Bytes(), &entry)
					Expect(entry.Uuid).To(Equal(entryUUID))
					Expect(entry.Settings.Level).To(Equal("1"))
				})

				It("Decrement", func() {
					input := openapi.SpacedRepetitionEntryViewed{
						Action: spaced_repetition.ActionDecrement,
						Uuid:   entryUUID,
					}

					b, _ := json.Marshal(input)
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
					c = e.NewContext(req, rec)

					c.Set("loggedInUser", *user)
					c.SetPath(uri)
					created, _ := time.Parse(time.RFC3339, "2020-12-27T17:04:59Z")
					whenNext, _ := time.Parse(time.RFC3339, "2020-12-27T18:04:59Z")
					repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{
						UUID:     entryUUID,
						UserUUID: userUUID,
						Body:     `{"show":"Hello","kind":"v1","uuid":"ba9277fc4c6190fb875ad8f9cee848dba699937f","data":"Hello","settings":{"level":"1","when_next":"2020-12-27T18:04:59Z","created":"2020-12-27T17:04:59Z"}}`,
						Created:  created,
						WhenNext: whenNext,
					}, nil)

					repo.On("UpdateEntry", mock.Anything).Return(nil)
					eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.ApiSpacedRepetition))

						b, _ := json.Marshal(moment.Data)
						var data spaced_repetition.EventSpacedRepetition
						json.Unmarshal(b, &data)

						Expect(data.Kind).To(Equal(spaced_repetition.EventKindViewed))
						Expect(data.Data.UserUUID).To(Equal(userUUID))
						Expect(data.Data.UUID).To(Equal(entryUUID))

						return true
					}))

					service.EntryViewed(c)
					Expect(rec.Code).To(Equal(http.StatusOK))
					var entry openapi.SpacedRepetitionV1
					json.Unmarshal(rec.Body.Bytes(), &entry)
					Expect(entry.Uuid).To(Equal(entryUUID))
					Expect(entry.Settings.Level).To(Equal("0"))
				})

				It("Decrement V2", func() {
					input := openapi.SpacedRepetitionEntryViewed{
						Action: spaced_repetition.ActionDecrement,
						Uuid:   "75698c0f5a7b904f1799ceb68e2afe67ad987689",
					}

					b, _ := json.Marshal(input)
					req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, string(b))
					c = e.NewContext(req, rec)

					c.Set("loggedInUser", *user)
					c.SetPath(uri)
					created, _ := time.Parse(time.RFC3339, "2020-12-28T11:44:33Z")
					whenNext, _ := time.Parse(time.RFC3339, "2020-12-28T12:44:33Z")
					repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{
						UUID:     entryUUIDV2,
						UserUUID: userUUID,
						Body:     `{"data":{"from":"March","to":"Mars"},"kind":"v2","settings":{"created":"2020-12-28T11:44:33Z","level":"0","show":"to","when_next":"2020-12-28T12:44:33Z"},"show":"Mars","uuid":"75698c0f5a7b904f1799ceb68e2afe67ad987689"}`,
						Created:  created,
						WhenNext: whenNext,
					}, nil)

					repo.On("UpdateEntry", mock.Anything).Return(nil)
					eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
						Expect(moment.Kind).To(Equal(event.ApiSpacedRepetition))

						b, _ := json.Marshal(moment.Data)
						var data spaced_repetition.EventSpacedRepetition
						json.Unmarshal(b, &data)

						Expect(data.Kind).To(Equal(spaced_repetition.EventKindViewed))
						Expect(data.Data.UserUUID).To(Equal(userUUID))
						Expect(data.Data.UUID).To(Equal(entryUUIDV2))

						return true
					}))

					service.EntryViewed(c)
					Expect(rec.Code).To(Equal(http.StatusOK))
					var entry openapi.SpacedRepetitionV2
					json.Unmarshal(rec.Body.Bytes(), &entry)
					Expect(entry.Uuid).To(Equal(entryUUIDV2))
					Expect(entry.Kind).To(Equal(alist.FromToList))
					Expect(entry.Show).To(Equal("Mars"))
					Expect(entry.Data.From).To(Equal("March"))
					Expect(entry.Data.To).To(Equal("Mars"))
					Expect(entry.Settings.Level).To(Equal("0"))
				})
			})
		})
	})

	When("Getting all entries", func() {
		var (
			uri = "/api/v1/api/v1/spaced-repetition/all"
		)

		It("Error looking up records", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)

			repo.On("GetEntries", user.Uuid).Return(nil, want)

			service.GetAll(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Records found, but empty", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)

			repo.On("GetEntries", user.Uuid).Return(make([]interface{}, 0), nil)

			service.GetAll(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
		})
	})

	When("Looking up the next entry", func() {
		var (
			uri = "/api/v1/api/v1/spaced-repetition/next"
		)

		It("Not found, meaning the user has not added any entries", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)

			repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{}, spaced_repetition.ErrNotFound)

			service.GetNext(c)
			Expect(rec.Code).To(Equal(http.StatusNotFound))
		})

		It("Found but in the future", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)
			whenNext := time.Now().UTC().Add(1 * time.Hour)

			repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{
				WhenNext: whenNext,
			}, nil)

			service.GetNext(c)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
		})

		It("An error looking up in repo", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)

			repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{}, want)
			service.GetNext(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})

		It("Entry found and ready", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodGet, uri, "")
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath(uri)

			repo.On("GetNext", user.Uuid).Return(spaced_repetition.SpacedRepetitionEntry{
				WhenNext: time.Now().UTC().Add(-1 * time.Hour),
				Body:     `{"data":"Hello","kind":"v1","settings":{"created":"2020-12-27T16:57:31Z","level":"0","when_next":"2020-12-27T17:57:31Z"},"show":"Hello","uuid":"ba9277fc4c6190fb875ad8f9cee848dba699937f"}`,
			}, nil)

			service.GetNext(c)
			Expect(rec.Code).To(Equal(http.StatusOK))
			var entry openapi.SpacedRepetitionV1
			json.Unmarshal(rec.Body.Bytes(), &entry)
			Expect(entry.Uuid).To(Equal(entryUUID))
		})
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
		)

		It("Missing the entry id", func() {
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

			repo.On("GetEntry", user.Uuid, entryUUID).Return(nil, want)
			service.DeleteEntry(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Entry failed due to repo delete", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, input)
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath("/api/v1/api/v1/spaced-repetition/:uuid")
			c.SetParamNames("uuid")
			c.SetParamValues(entryUUID)

			repo.On("GetEntry", user.Uuid, entryUUID).Return(nil, nil)
			repo.On("DeleteEntry", user.Uuid, entryUUID).Return(want)
			service.DeleteEntry(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
			testutils.CheckMessageResponseFromResponseRecorder(rec, i18n.InternalServerErrorFunny)
		})

		It("Entry deleted", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodDelete, uri, input)
			c = e.NewContext(req, rec)

			c.Set("loggedInUser", *user)
			c.SetPath("/api/v1/api/v1/spaced-repetition/:uuid")
			c.SetParamNames("uuid")
			c.SetParamValues(entryUUID)

			repo.On("GetEntry", user.Uuid, entryUUID).Return(nil, nil)
			repo.On("DeleteEntry", user.Uuid, entryUUID).Return(nil)
			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(event.ApiSpacedRepetition))
				b, _ := json.Marshal(moment.Data)
				var data spaced_repetition.EventSpacedRepetition
				json.Unmarshal(b, &data)

				Expect(data.Kind).To(Equal(spaced_repetition.EventKindDeleted))
				Expect(data.Data.UserUUID).To(Equal(userUUID))
				Expect(data.Data.UUID).To(Equal(entryUUID))
				return true
			}))

			service.DeleteEntry(c)
			Expect(rec.Code).To(Equal(http.StatusNoContent))
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

		It("New Entry V1", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			repo.On("SaveEntry", mock.Anything).Return(nil)

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

		It("New Entry V2", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, inputV2)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			repo.On("SaveEntry", mock.Anything).Return(nil)

			eventMessageBus.On("Publish", event.TopicMonolog, mock.MatchedBy(func(moment event.Eventlog) bool {
				Expect(moment.Kind).To(Equal(event.ApiSpacedRepetition))
				b, _ := json.Marshal(moment.Data)
				var data spaced_repetition.EventSpacedRepetition
				json.Unmarshal(b, &data)

				Expect(data.Kind).To(Equal(spaced_repetition.EventKindNew))
				Expect(data.Data.UserUUID).To(Equal(userUUID))
				Expect(data.Data.UUID).To(Equal("75698c0f5a7b904f1799ceb68e2afe67ad987689"))
				return true
			}))

			service.SaveEntry(c)
			Expect(rec.Code).To(Equal(http.StatusCreated))
			var entry openapi.SpacedRepetitionV2
			json.Unmarshal(rec.Body.Bytes(), &entry)

			Expect(entry.Uuid).To(Equal(entryUUIDV2))
			Expect(entry.Kind).To(Equal(alist.FromToList))
			Expect(entry.Show).To(Equal("Mars"))
			Expect(entry.Data.From).To(Equal("March"))
			Expect(entry.Data.To).To(Equal("Mars"))
			Expect(entry.Settings.Level).To(Equal("0"))
		})

		It("Failed to save entry", func() {
			req, rec = testutils.SetupJSONEndpoint(http.MethodPost, uri, input)
			c = e.NewContext(req, rec)
			c.SetPath(uri)
			c.Set("loggedInUser", *user)

			repo.On("SaveEntry", mock.Anything).Return(want)
			service.SaveEntry(c)
			Expect(rec.Code).To(Equal(http.StatusInternalServerError))
		})
	})
})
