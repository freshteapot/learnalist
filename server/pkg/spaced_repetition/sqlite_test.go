package spaced_repetition_test

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced Repetitiion Repository Sqlite", func() {
	var (
		dbCon               *sqlx.DB
		mockSql             sqlmock.Sqlmock
		repo                spaced_repetition.SpacedRepetitionRepository
		entryUUID, userUUID string
		want                error
		entry               spaced_repetition.SpacedRepetitionEntry
		created, whenNext   time.Time
	)

	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()
		repo = spaced_repetition.NewSqliteRepository(dbCon)
		entryUUID = "ba9277fc4c6190fb875ad8f9cee848dba699937f"
		userUUID = "fake-user-123"

		want = errors.New("fail")
		created, _ = time.Parse(time.RFC3339, "2020-12-27T17:04:59Z")
		whenNext, _ = time.Parse(time.RFC3339, "2020-12-27T18:04:59Z")
		entry = spaced_repetition.SpacedRepetitionEntry{
			UUID:     entryUUID,
			UserUUID: userUUID,
			Body:     `{"show":"Hello","kind":"v1","uuid":"ba9277fc4c6190fb875ad8f9cee848dba699937f","data":"Hello","settings":{"level":"0","when_next":"2020-12-27T18:04:59Z","created":"2020-12-27T17:04:59Z"}}`,
			Created:  created,
			WhenNext: whenNext,
		}
	})

	AfterEach(func() {
		dbCon.Close()
	})

	When("Saving", func() {
		var (
			entry     spaced_repetition.SpacedRepetitionEntry
			whenNext  time.Time
			created   time.Time
			sqlExpect *sqlmock.ExpectedExec
		)

		BeforeEach(func() {
			created, _ = time.Parse(time.RFC3339, "2020-12-27T17:04:59Z")
			whenNext, _ = time.Parse(time.RFC3339, "2020-12-27T18:04:59Z")
			entry = spaced_repetition.SpacedRepetitionEntry{
				UUID:     entryUUID,
				UserUUID: userUUID,
				Body:     `{"show":"Hello","kind":"v1","uuid":"ba9277fc4c6190fb875ad8f9cee848dba699937f","data":"Hello","settings":{"level":"0","when_next":"2020-12-27T18:04:59Z","created":"2020-12-27T17:04:59Z"}}`,
				Created:  created,
				WhenNext: whenNext,
			}

			sqlExpect = mockSql.ExpectExec(spaced_repetition.SqlSaveItem)
		})

		It("Fail to save", func() {
			sqlExpect.WillReturnError(want).WithArgs(
				entry.UUID,
				entry.Body,
				entry.UserUUID,
				entry.WhenNext.Format(time.RFC3339),
				entry.Created.Format(time.RFC3339),
			)
			err := repo.SaveEntry(entry)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("already in the system", func() {
			want = errors.New("UNIQUE constraint failed XXX")
			sqlExpect.WillReturnError(want).WithArgs(
				entry.UUID,
				entry.Body,
				entry.UserUUID,
				entry.WhenNext.Format(time.RFC3339),
				entry.Created.Format(time.RFC3339),
			)
			err := repo.SaveEntry(entry)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(spaced_repetition.ErrSpacedRepetitionEntryExists))
		})

		It("Success", func() {
			sqlExpect.WillReturnResult(sqlmock.NewResult(1, 1)).WithArgs(
				entry.UUID,
				entry.Body,
				entry.UserUUID,
				entry.WhenNext.Format(time.RFC3339),
				entry.Created.Format(time.RFC3339),
			)
			err := repo.SaveEntry(entry)
			Expect(err).To(BeNil())
		})
	})

	When("Deleting", func() {
		It("Fail", func() {
			mockSql.ExpectExec(spaced_repetition.SqlDeleteItem).WillReturnError(want)
			err := repo.DeleteEntry(userUUID, entryUUID)
			Expect(err).To(Equal(want))
		})
		It("Success", func() {
			mockSql.ExpectExec(spaced_repetition.SqlDeleteItem).
				WillReturnResult(sqlmock.NewResult(1, 1))
			err := repo.DeleteEntry(userUUID, entryUUID)
			Expect(err).To(BeNil())
		})
	})

	When("Deleting by user", func() {
		It("Fail", func() {
			mockSql.ExpectExec(spaced_repetition.SqlDeleteByUser).
				WithArgs(userUUID).
				WillReturnError(want)
			err := repo.DeleteByUser(userUUID)
			Expect(err).To(Equal(want))
		})
		It("Success", func() {
			mockSql.ExpectExec(spaced_repetition.SqlDeleteByUser).
				WithArgs(userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))
			err := repo.DeleteByUser(userUUID)
			Expect(err).To(BeNil())
		})
	})

	When("Updating", func() {
		It("Fail", func() {
			mockSql.ExpectExec(spaced_repetition.SqlUpdateItem).
				WillReturnError(want).
				WithArgs(entry.Body, entry.WhenNext.Format(time.RFC3339), entry.UserUUID, entry.UUID)
			err := repo.UpdateEntry(entry)
			Expect(err).To(Equal(want))
		})

		It("Success", func() {
			mockSql.ExpectExec(spaced_repetition.SqlUpdateItem).
				WillReturnResult(sqlmock.NewResult(1, 1)).
				WithArgs(entry.Body, entry.WhenNext.Format(time.RFC3339), entry.UserUUID, entry.UUID)

			err := repo.UpdateEntry(entry)
			Expect(err).To(BeNil())
		})
	})

	When("Getting all entries", func() {
		It("Fail", func() {
			mockSql.ExpectQuery(spaced_repetition.SqlGetAll).
				WillReturnError(want).
				WithArgs(userUUID)
			_, err := repo.GetEntries(userUUID)
			Expect(err).To(Equal(want))
		})
		It("Success", func() {
			rs := sqlmock.NewRows([]string{
				"body",
			}).
				AddRow(entry.Body)

			mockSql.ExpectQuery(spaced_repetition.SqlGetAll).
				WithArgs(userUUID).
				WillReturnRows(rs)
			items, err := repo.GetEntries(userUUID)
			Expect(err).To(BeNil())
			var entry openapi.SpacedRepetitionV1
			b, _ := json.Marshal(items[0])
			json.Unmarshal(b, &entry)
			Expect(entry.Kind).To(Equal(alist.SimpleList))
		})
	})

	When("Getting entry by user and uuid", func() {
		It("Fail", func() {
			mockSql.ExpectQuery(spaced_repetition.SqlGetItem).
				WillReturnError(want).
				WithArgs(entryUUID, userUUID)
			_, err := repo.GetEntry(userUUID, entryUUID)
			Expect(err).To(Equal(want))
		})

		It("Return correct not found", func() {
			mockSql.ExpectQuery(spaced_repetition.SqlGetItem).
				WillReturnError(sql.ErrNoRows).
				WithArgs(entryUUID, userUUID)
			_, err := repo.GetEntry(userUUID, entryUUID)
			Expect(err).To(Equal(spaced_repetition.ErrNotFound))
		})

		It("Success", func() {
			rs := sqlmock.NewRows([]string{
				"uuid",
				"body",
				"user_uuid",
				"when_next",
				"created",
			}).
				AddRow(entry.UUID, entry.Body, entry.UserUUID, whenNext, created)

			mockSql.ExpectQuery(spaced_repetition.SqlGetItem).
				WithArgs(entryUUID, userUUID).
				WillReturnRows(rs)
			item, err := repo.GetEntry(userUUID, entryUUID)
			Expect(err).To(BeNil())
			var entry openapi.SpacedRepetitionV1
			b, _ := json.Marshal(item)
			json.Unmarshal(b, &entry)
			Expect(entry.Kind).To(Equal(alist.SimpleList))
			Expect(entry.Uuid).To(Equal(entryUUID))
		})
	})

	When("Getting next entry for a user", func() {
		It("Fail to lookup via the repo", func() {
			mockSql.ExpectQuery(spaced_repetition.SqlGetNext).
				WillReturnError(want).
				WithArgs(userUUID)
			_, err := repo.GetNext(userUUID)
			Expect(err).To(Equal(want))
		})

		It("Return correct not found", func() {
			mockSql.ExpectQuery(spaced_repetition.SqlGetNext).
				WillReturnError(sql.ErrNoRows).
				WithArgs(userUUID)
			_, err := repo.GetNext(userUUID)
			Expect(err).To(Equal(spaced_repetition.ErrNotFound))
		})

		It("Success", func() {
			rs := sqlmock.NewRows([]string{
				"uuid",
				"body",
				"user_uuid",
				"when_next",
				"created",
			}).
				AddRow(entry.UUID, entry.Body, entry.UserUUID, whenNext, created)

			mockSql.ExpectQuery(spaced_repetition.SqlGetNext).
				WithArgs(userUUID).
				WillReturnRows(rs)
			item, err := repo.GetNext(userUUID)
			Expect(err).To(BeNil())
			var entry openapi.SpacedRepetitionV1
			json.Unmarshal([]byte(item.Body), &entry)
			Expect(entry.Kind).To(Equal(alist.SimpleList))
			Expect(entry.Uuid).To(Equal(entryUUID))
		})
	})
})
