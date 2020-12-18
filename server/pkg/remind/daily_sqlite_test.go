package remind_test

import (
	"encoding/json"
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/pkg/apps"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/remind"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Daily Sqlite", func() {
	var (
		dbCon   *sqlx.DB
		mockSql sqlmock.Sqlmock
		repo    remind.RemindDailySettingsRepository
	)

	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()
		repo = remind.NewRemindDailySettingsSqliteRepository(dbCon)
	})

	AfterEach(func() {
		dbCon.Close()
	})

	When("Saving", func() {
		var (
			want      error
			userUUID  string
			settings  openapi.RemindDailySettings
			whenNext  string
			body      string
			sqlExpect *sqlmock.ExpectedExec
		)
		BeforeEach(func() {
			want = errors.New("fail")
			userUUID = "fake-user-123"
			settings = openapi.RemindDailySettings{
				AppIdentifier: apps.RemindV1,
			}
			whenNext = "2020-12-15T14:30:30Z"
			b, _ := json.Marshal(settings)
			body = string(b)
			sqlExpect = mockSql.ExpectExec(remind.SqlSave)

		})
		It("Fail to save", func() {
			sqlExpect.WillReturnError(want).WithArgs(
				userUUID,
				settings.AppIdentifier,
				body,
				whenNext,
				body,
				whenNext,
			)
			err := repo.Save(userUUID, settings, whenNext)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
		})
		It("Success", func() {
			sqlExpect.WillReturnResult(sqlmock.NewResult(1, 1))
			err := repo.Save(userUUID, settings, whenNext)
			Expect(err).To(BeNil())
		})
	})

	When("Deleting", func() {
		var (
			want          error
			userUUID      string
			appIdentifier string
		)
		BeforeEach(func() {
			want = errors.New("fail")
			userUUID = "fake-user-123"
			appIdentifier = apps.RemindV1
		})

		By("via user", func() {
			It("fail", func() {
				mockSql.ExpectExec(remind.SqlDeleteByUser).WillReturnError(want)
				err := repo.DeleteByUser(userUUID)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("success", func() {
				mockSql.ExpectExec(remind.SqlDeleteByUser).WillReturnResult(sqlmock.NewResult(1, 1))
				err := repo.DeleteByUser(userUUID)
				Expect(err).To(BeNil())
			})
		})

		By("via app", func() {
			It("fail", func() {
				mockSql.ExpectExec(remind.SqlDeleteByDeviceInfo).WillReturnError(want).WithArgs(
					userUUID,
					appIdentifier,
				)
				err := repo.DeleteByApp(userUUID, appIdentifier)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(want))
			})

			It("success", func() {
				mockSql.ExpectExec(remind.SqlDeleteByDeviceInfo).WillReturnResult(sqlmock.NewResult(1, 1))
				err := repo.DeleteByApp(userUUID, appIdentifier)
				Expect(err).To(BeNil())
			})
		})
	})
})
