package sqlite_test

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/label"
	storage "github.com/freshteapot/learnalist-api/server/api/label/sqlite"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Label", func() {
	var (
		dbCon   *sqlx.DB
		mockSql sqlmock.Sqlmock
		labels  label.LabelReadWriter
	)

	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()
		labels = storage.NewLabel(dbCon)
	})

	AfterEach(func() {
		dbCon.Close()
	})

	It("", func() {
		fmt.Println(mockSql, labels)
		Expect("").To(Equal(""))
	})

	When("Posting a list label", func() {
		var (
			input label.AlistLabel
		)
		BeforeEach(func() {
			input = label.NewAlistLabel("test", "fake-user-123", "fake-list-123")
		})

		It("Not valid", func() {
			input.Label = ""
			resp, err := labels.PostAlistLabel(input)
			Expect(err).To(HaveOccurred())
			Expect(resp).To(Equal(http.StatusBadRequest))
		})

		It("Success", func() {
			mockSql.ExpectExec(storage.SqlInserListLabel).
				WithArgs(input.Label, input.UserUuid, input.AlistUuid).
				WillReturnResult(sqlmock.NewResult(1, 1))

			resp, err := labels.PostAlistLabel(input)
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(http.StatusCreated))
		})

		It("Duplicate label", func() {
			By("First add label")
			mockSql.ExpectExec(storage.SqlInserListLabel).
				WithArgs(input.Label, input.UserUuid, input.AlistUuid).
				WillReturnResult(sqlmock.NewResult(1, 1))

			resp, err := labels.PostAlistLabel(input)
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(http.StatusCreated))

			By("Confirm adding again fails")
			want := errors.New("UNIQUE constraint failed")
			mockSql.ExpectExec(storage.SqlInserListLabel).WillReturnError(want)
			resp, err = labels.PostAlistLabel(input)
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(http.StatusOK))
		})

		It("DB issue", func() {
			want := errors.New("I must fail")
			mockSql.ExpectExec(storage.SqlInserListLabel).WillReturnError(want)
			resp, err := labels.PostAlistLabel(input)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
			Expect(resp).To(Equal(http.StatusInternalServerError))
		})
	})

	When("Posting a user label", func() {
		var (
			input label.UserLabel
		)
		BeforeEach(func() {
			input = label.NewUserLabel("test", "fake-user-123")
		})

		It("Not valid", func() {
			input.Label = ""
			resp, err := labels.PostUserLabel(input)
			Expect(err).To(HaveOccurred())
			Expect(resp).To(Equal(http.StatusBadRequest))
		})

		It("Success", func() {
			mockSql.ExpectExec(storage.SqlInserUserLabel).
				WithArgs(input.Label, input.UserUuid).
				WillReturnResult(sqlmock.NewResult(1, 1))

			resp, err := labels.PostUserLabel(input)
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(http.StatusCreated))
		})

		It("Duplicate label", func() {
			By("First add label")
			mockSql.ExpectExec(storage.SqlInserUserLabel).
				WithArgs(input.Label, input.UserUuid).
				WillReturnResult(sqlmock.NewResult(1, 1))

			resp, err := labels.PostUserLabel(input)
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(http.StatusCreated))

			By("Confirm adding again fails")
			want := errors.New("UNIQUE constraint failed")
			mockSql.ExpectExec(storage.SqlInserUserLabel).WillReturnError(want)
			resp, err = labels.PostUserLabel(input)
			Expect(err).To(BeNil())
			Expect(resp).To(Equal(http.StatusOK))
		})

		It("DB issue", func() {
			want := errors.New("I must fail")
			mockSql.ExpectExec(storage.SqlInserUserLabel).WillReturnError(want)
			resp, err := labels.PostUserLabel(input)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
			Expect(resp).To(Equal(http.StatusInternalServerError))
		})
	})

	When("Looking for lists", func() {
		var (
			label    = "water"
			userUUID = "fake-123"
		)

		It("When there is an error", func() {
			want := sql.ErrNoRows

			mockSql.ExpectQuery(storage.SqlGetListsByUserAndLabel).WithArgs(userUUID, label).
				WillReturnError(want)

			resp, err := labels.GetUniqueListsByUserAndLabel(label, userUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
			Expect(len(resp)).To(Equal(0))
		})

		It("Lists found", func() {
			lists := []string{"fake-123", "fake-456"}
			rs := sqlmock.NewRows([]string{
				"alist_uuid",
			}).
				AddRow(lists[0]).
				AddRow(lists[1])

			mockSql.ExpectQuery(storage.SqlGetListsByUserAndLabel).WithArgs(userUUID, label).
				WillReturnRows(rs)

			resp, err := labels.GetUniqueListsByUserAndLabel(label, userUUID)
			Expect(err).To(BeNil())
			Expect(len(resp)).To(Equal(2))
			Expect(resp).To(Equal(lists))
		})
	})

	When("Getting Labels for a user", func() {
		var (
			userUUID = "fake-123"
		)

		It("When there is an error", func() {
			want := sql.ErrNoRows

			mockSql.ExpectQuery(storage.SqlGetUserLabels).WithArgs(userUUID, userUUID).
				WillReturnError(want)

			resp, err := labels.GetUserLabels(userUUID)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
			Expect(len(resp)).To(Equal(0))
		})

		It("Labels found", func() {
			items := []string{"label-123", "label-456"}
			rs := sqlmock.NewRows([]string{
				"alist_uuid",
			}).
				AddRow(items[0]).
				AddRow(items[1])

			mockSql.ExpectQuery(storage.SqlGetUserLabels).WithArgs(userUUID, userUUID).
				WillReturnRows(rs)

			resp, err := labels.GetUserLabels(userUUID)
			Expect(err).To(BeNil())
			Expect(len(resp)).To(Equal(2))
			Expect(resp).To(Equal(items))
		})
	})
	When("Removing", func() {
		When("Labels from a list", func() {
			var (
				alistUUID = "fake-list-123"
			)
			Specify("List is empty", func() {
				err := labels.RemoveLabelsForAlist("")
				Expect(err).To(BeNil())
			})

			Specify("Sucessfully removed", func() {
				mockSql.ExpectBegin()
				mockSql.ExpectExec(storage.SqlDeleteLabelByList).
					WithArgs(alistUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit()

				err := labels.RemoveLabelsForAlist(alistUUID)
				Expect(err).To(BeNil())
			})

			Specify("Error whilst removing", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(storage.SqlDeleteLabelByList).
					WithArgs(alistUUID).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit().WillReturnError(want)

				err := labels.RemoveLabelsForAlist(alistUUID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})

		When("Removing labels from a user", func() {
			var (
				label    = "water"
				userUUID = "fake-123"
			)

			Specify("Sucessfully removed", func() {
				mockSql.ExpectBegin()
				mockSql.ExpectExec(storage.SqlDeleteLabelByUser).
					WithArgs(userUUID, label).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(storage.SqlDeleteLabelByUserFromList).
					WithArgs(userUUID, label).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit()

				err := labels.RemoveUserLabel(label, userUUID)
				Expect(err).To(BeNil())
			})

			Specify("Error whilst removing", func() {
				want := errors.New("sql: TX")
				mockSql.ExpectBegin()
				mockSql.ExpectExec(storage.SqlDeleteLabelByUser).
					WithArgs(userUUID, label).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectExec(storage.SqlDeleteLabelByUserFromList).
					WithArgs(userUUID, label).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mockSql.ExpectCommit().WillReturnError(want)

				err := labels.RemoveUserLabel(label, userUUID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})
		})
	})
})
