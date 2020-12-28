package remind_test

import (
	"errors"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Spaced Repetition Sqlite", func() {
	var (
		dbCon   *sqlx.DB
		mockSql sqlmock.Sqlmock

		userUUID string
		want     error
		repo     remind.RemindSpacedRepetitionRepository
	)

	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()

		userUUID = "fake-user-123"
		want = errors.New("fail")
		repo = remind.NewRemindSpacedRepetitionSqliteRepository(dbCon)
	})

	AfterEach(func() {
		dbCon.Close()
	})

	When("Getting reminders", func() {
		It("Issue talking to the database", func() {
			whenNext, lastActive := remind.DefaultWhenNextWithLastActiveOffset()
			mockSql.ExpectQuery(remind.SpacedRepetitionSqlGetReminders).
				WithArgs(whenNext, lastActive).
				WillReturnError(want)
			_, err := repo.GetReminders(whenNext, lastActive)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("success, nothing found", func() {
			whenNext, lastActive := remind.DefaultWhenNextWithLastActiveOffset()

			rs := sqlmock.NewRows([]string{"user_uuid", "when_next", "last_active", "medium"})

			mockSql.ExpectQuery(remind.SpacedRepetitionSqlGetReminders).
				WithArgs(whenNext, lastActive).
				WillReturnRows(rs)

			items, err := repo.GetReminders(whenNext, lastActive)
			Expect(err).To(BeNil())
			Expect(len(items)).To(Equal(0))
		})

		It("success, record found", func() {
			whenNext, lastActive := remind.DefaultWhenNextWithLastActiveOffset()
			tWhenNext, _ := time.Parse(time.RFC3339Nano, whenNext)
			tLastActive, _ := time.Parse(time.RFC3339Nano, whenNext)

			rs := sqlmock.NewRows([]string{"user_uuid", "when_next", "last_active", "medium"}).
				AddRow("fake-user-123", tWhenNext, tLastActive, "").
				AddRow("fake-user-456", tWhenNext, tLastActive, "fake-token")

			mockSql.ExpectQuery(remind.SpacedRepetitionSqlGetReminders).
				WithArgs(whenNext, lastActive).
				WillReturnRows(rs)

			items, err := repo.GetReminders(whenNext, lastActive)
			Expect(err).To(BeNil())
			Expect(len(items)).To(Equal(2))
			Expect(items[0].UserUUID).To(Equal("fake-user-123"))
			Expect(items[0].Medium).To(Equal(""))
			Expect(items[1].UserUUID).To(Equal("fake-user-456"))
			Expect(items[1].Medium).To(Equal("fake-token"))
		})
	})
	When("Setting a reminder", func() {
		var (
			whenNext, lastActive   time.Time
			sWhenNext, sLastActive string
		)

		BeforeEach(func() {
			whenNext = time.Now().UTC()
			lastActive = time.Now().UTC()
			sWhenNext = whenNext.Format(time.RFC3339)
			sLastActive = lastActive.Format(time.RFC3339)
		})

		It("fail", func() {
			mockSql.ExpectExec(remind.SpacedRepetitionSqlSave).
				WithArgs(
					userUUID, sWhenNext, sLastActive, // New
					sWhenNext, sLastActive, // On conflict
				).
				WillReturnError(want)

			err := repo.SetReminder(userUUID, whenNext, lastActive)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("success", func() {
			mockSql.ExpectExec(remind.SpacedRepetitionSqlSave).
				WithArgs(
					userUUID, sWhenNext, sLastActive, // New
					sWhenNext, sLastActive, // On conflict
				).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.SetReminder(userUUID, whenNext, lastActive)
			Expect(err).To(BeNil())
		})
	})

	When("Deleting", func() {
		It("fail", func() {
			mockSql.ExpectExec(remind.SpacedRepetitionSqlDeleteByUser).
				WillReturnError(want).
				WithArgs(userUUID)
			err := repo.DeleteByUser(userUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("success", func() {
			mockSql.ExpectExec(remind.SpacedRepetitionSqlDeleteByUser).
				WithArgs(userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))
			err := repo.DeleteByUser(userUUID)
			Expect(err).To(BeNil())
		})
	})

	When("Update the sent status", func() {
		It("fail", func() {
			mockSql.ExpectExec(remind.SpacedRepetitionSqlUpdateSent).
				WithArgs(remind.ReminderNotSentYet, userUUID).
				WillReturnError(want)

			err := repo.UpdateSent(userUUID, remind.ReminderNotSentYet)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})

		It("success", func() {
			mockSql.ExpectExec(remind.SpacedRepetitionSqlUpdateSent).
				WithArgs(remind.ReminderNotSentYet, userUUID).
				WillReturnResult(sqlmock.NewResult(1, 1))

			err := repo.UpdateSent(userUUID, remind.ReminderNotSentYet)
			Expect(err).To(BeNil())
		})
	})
})
