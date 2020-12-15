package challenge_test

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/label"
	"github.com/freshteapot/learnalist-api/server/pkg/challenge"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"

	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Sqlite", func() {
	var (
		dbCon         *sqlx.DB
		mockSql       sqlmock.Sqlmock
		labels        label.LabelReadWriter
		challengeRepo challenge.ChallengeRepository
	)

	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()
		challengeRepo = challenge.NewSqliteRepository(dbCon)
	})

	AfterEach(func() {
		dbCon.Close()
	})

	It("", func() {
		fmt.Println(mockSql, labels)
		Expect("").To(Equal(""))
	})

	It("", func() {
		var err error
		Expect(err).ToNot(HaveOccurred())
	})

	When("Getting challenges for a user", func() {
		var (
			kind       = ""
			userUUID   = "fake-123"
			preQuery   = fmt.Sprintf(challenge.SqlGetChallengesByUser, userUUID, userUUID)
			challenge1 challenge.ChallengeShortInfo
			challenge2 challenge.ChallengeShortInfo
			created    time.Time
		)

		BeforeEach(func() {
			created, _ = time.Parse(time.RFC3339Nano, "2020-12-15T14:30:30Z")
			challenge1 = challenge.ChallengeShortInfo{
				UUID:        "fake-challenge-1",
				Kind:        challenge.KindPlankGroup,
				Description: "hello",
				Created:     created.Format(time.RFC3339Nano),
				CreatedBy:   "fake-user-1",
			}
			challenge2 = challenge.ChallengeShortInfo{
				UUID:        "fake-challenge-2",
				Kind:        "todo",
				Description: "hello",
				Created:     created.Format(time.RFC3339Nano),
				CreatedBy:   "fake-user-1",
			}
		})

		It("When there is an error", func() {
			want := sql.ErrNoRows
			query, args, _ := sqlx.In(preQuery, userUUID, challenge.ChallengeKinds)
			query = dbCon.Rebind(query)

			// TODO if we add new challenge kinds, this will break
			mockSql.ExpectQuery(query).WithArgs(args[0], args[1], args[2]).
				WillReturnError(want)

			resp, err := challengeRepo.GetChallengesByUser(userUUID, kind)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(want))
			Expect(len(resp)).To(Equal(0))
		})

		It("no results", func() {
			query, args, _ := sqlx.In(preQuery, userUUID, []string{challenge.KindPlankGroup})
			query = dbCon.Rebind(query)

			mockSql.ExpectQuery(query).WithArgs(args[0], args[1]).WillReturnError(sql.ErrNoRows)

			resp, err := challengeRepo.GetChallengesByUser(userUUID, challenge.KindPlankGroup)
			Expect(err).To(Equal(sql.ErrNoRows))
			Expect(len(resp)).To(Equal(0))
		})

		It("Find all challenges", func() {
			rs := sqlmock.NewRows([]string{
				"uuid",
				"kind",
				"description",
				"created",
				"user_uuid",
			}).
				AddRow(challenge1.UUID, challenge1.Kind, challenge1.Description, created, challenge1.CreatedBy).
				AddRow(challenge2.UUID, challenge2.Kind, challenge2.Description, created, challenge2.CreatedBy)

			query, args, _ := sqlx.In(preQuery, userUUID, challenge.ChallengeKinds)
			query = dbCon.Rebind(query)
			mockSql.ExpectQuery(query).WithArgs(args[0], args[1], args[2]).WillReturnRows(rs)

			resp, err := challengeRepo.GetChallengesByUser(userUUID, "")
			Expect(err).To(BeNil())
			Expect(len(resp)).To(Equal(2))
			Expect(resp[0]).To(Equal(challenge1))
		})

		It("Find only plank-group", func() {
			rs := sqlmock.NewRows([]string{
				"uuid",
				"kind",
				"description",
				"created",
				"user_uuid",
			}).
				AddRow(challenge1.UUID, challenge1.Kind, challenge1.Description, created, challenge1.CreatedBy)

			query, args, _ := sqlx.In(preQuery, userUUID, []string{challenge.KindPlankGroup})
			query = dbCon.Rebind(query)
			mockSql.ExpectQuery(query).WithArgs(args[0], args[1]).WillReturnRows(rs)

			resp, err := challengeRepo.GetChallengesByUser(userUUID, challenge.KindPlankGroup)
			Expect(err).To(BeNil())
			Expect(len(resp)).To(Equal(1))
			Expect(resp[0]).To(Equal(challenge1))
		})
	})
})
