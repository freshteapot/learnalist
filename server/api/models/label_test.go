package models_test

/*
import (
	"errors"
	"fmt"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/i18n"
	labelStorage "github.com/freshteapot/learnalist-api/server/api/label/sqlite"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/mocks"
	aclStorage "github.com/freshteapot/learnalist-api/server/pkg/acl/sqlite"
	oauthStorage "github.com/freshteapot/learnalist-api/server/pkg/oauth/sqlite"
	helper "github.com/freshteapot/learnalist-api/server/pkg/testhelper"
	userStorage "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/jmoiron/sqlx"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
)

var _ = Describe("Testing Label", func() {
	// Not done
	// Possible issue with SaveList
	// Could ignore and move that
	var (
		datastore *mocks.Datastore
		dbCon     *sqlx.DB
		mockSql   sqlmock.Sqlmock
	)
	BeforeEach(func() {
		dbCon, mockSql, _ = helper.GetMockDB()
		fmt.Println(mockSql)
		acl := aclStorage.NewAcl(dbCon)
		userSession := userStorage.NewUserSession(dbCon)
		userFromIDP := userStorage.NewUserFromIDP(dbCon)
		userWithUsernameAndPassword := userStorage.NewUserWithUsernameAndPassword(dbCon)
		oauthHandler := oauthStorage.NewOAuthReadWriter(dbCon)
		labels := labelStorage.NewLabel(dbCon)
		dal = models.NewDAL(dbCon, acl, labels, userSession, userFromIDP, userWithUsernameAndPassword, oauthHandler)
	})

	It("", func() {
		datastore.On("SaveAlist", mock.Anything, mock.Anything).Return(alist.Alist{}, errors.New(i18n.InputSaveAlistOperationOwnerOnly))
		Expect("").To(Equal(""))
	})
})
*/
func (suite *ModelSuite) TestRemoveLabelsFromExistingLists() {
	setup := `
INSERT INTO user VALUES('c3d330fd-73b9-5d3c-840e-3bb59367b5ed','4187246584872952904','test1');
INSERT INTO alist_kv VALUES('3c9394eb-7df1-5611-bce1-2bc3a198b2a6','v1','{"data":[],"info":{"title":"Days of the Week","type":"v1","labels":["label a","label b"]},"uuid":"3c9394eb-7df1-5611-bce1-2bc3a198b2a6"}','c3d330fd-73b9-5d3c-840e-3bb59367b5ed');
INSERT INTO user_labels VALUES('label a','c3d330fd-73b9-5d3c-840e-3bb59367b5ed');
INSERT INTO user_labels VALUES('label b','c3d330fd-73b9-5d3c-840e-3bb59367b5ed');
INSERT INTO alist_labels VALUES('3c9394eb-7df1-5611-bce1-2bc3a198b2a6','c3d330fd-73b9-5d3c-840e-3bb59367b5ed','label a');
INSERT INTO alist_labels VALUES('3c9394eb-7df1-5611-bce1-2bc3a198b2a6','c3d330fd-73b9-5d3c-840e-3bb59367b5ed','label b');
`
	alist_uuid := "3c9394eb-7df1-5611-bce1-2bc3a198b2a6"
	user_uuid := "c3d330fd-73b9-5d3c-840e-3bb59367b5ed"
	label_remove := "label a"
	label_find := "label b"
	dal.Db.MustExec(setup)
	// Confirm the data is correct with two items for the list.
	alist, _ := dal.GetAlist(alist_uuid)
	suite.Equal(2, len(alist.Info.Labels))
	// Confirm the data is correct with two items.
	labels, _ := dal.Labels().GetUserLabels(user_uuid)
	suite.Equal(2, len(labels))

	err := dal.RemoveUserLabel(label_remove, user_uuid)

	suite.NoError(err)
	// Confirm the func returns the correct data.
	labels, _ = dal.Labels().GetUserLabels(user_uuid)
	suite.Equal(1, len(labels))
	suite.Equal(label_find, labels[0])
	// Confirm the list has been updated
	alist, _ = dal.GetAlist(alist_uuid)
	suite.Equal(1, len(alist.Info.Labels))
	suite.Equal(label_find, alist.Info.Labels[0])
}
