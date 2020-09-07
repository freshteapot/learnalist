package sqlite_test

import (
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

})
