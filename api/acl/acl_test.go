package acl

import (
	"fmt"
	"testing"

	"github.com/freshteapot/learnalist-api/api/alist"
	"github.com/freshteapot/learnalist-api/api/database"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

var db *sqlx.DB
var acl *Acl

type AclSuite struct {
	suite.Suite
}

func (suite *AclSuite) SetupSuite() {
	db = database.NewTestDB()
	acl = NewAclFromModel(database.PathToTestSqliteDb)
}

func (suite *AclSuite) SetupTest() {

}

func (suite *AclSuite) TearDownTest() {
	database.EmptyDatabase(db)
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(AclSuite))
}

func (suite *AclSuite) TestPublicWrite() {
	acl.createPublicRole()
	a := acl.enforcer.GetPolicy()
	policy := a[0]
	sub := policy[0]
	obj := policy[1]
	act := policy[2]
	suite.Equal("public:write", sub)
	suite.Equal("public", obj)
	suite.Equal("write", act)
}

func (suite *AclSuite) TestCreateListRole() {
	alistUUID := "fakeList123"
	acl.CreateListRole(alistUUID)

	filteredPolicy := acl.enforcer.GetFilteredPolicy(1, alistUUID)
	policyRead := filteredPolicy[0]
	policyReadSub := policyRead[0]
	policyReadObj := policyRead[1]
	policyReadAct := policyRead[2]

	policyWrite := filteredPolicy[1]
	policyWriteSub := policyWrite[0]
	policyWriteObj := policyWrite[1]
	policyWriteAct := policyWrite[2]

	read := fmt.Sprintf("%s:read", alistUUID)
	write := fmt.Sprintf("%s:write", alistUUID)

	suite.Equal(policyWriteSub, write)
	suite.Equal(policyWriteObj, alistUUID)
	suite.Equal(policyWriteAct, "write")

	suite.Equal(policyReadSub, read)
	suite.Equal(policyReadObj, alistUUID)
	suite.Equal(policyReadAct, "read")

	acl.DeleteListRole(alistUUID)
	filteredPolicy = acl.enforcer.GetFilteredPolicy(1, alistUUID)
	suite.Equal(0, len(filteredPolicy))
}

func (suite *AclSuite) TestGrantListPublicWriteAccess() {
	userUUID := "fakeUser123"
	acl.GrantListPublicWriteAccess(userUUID)
	roles := acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal("public:write", roles[0])
	suite.True(acl.HasUserPublicWriteAccess(userUUID))

	acl.RevokeListPublicWriteAccess(userUUID)
	suite.False(acl.HasUserPublicWriteAccess(userUUID))
}

func (suite *AclSuite) TestGrantAndRevokeListReadAccess() {
	userUUID := "fakeUser123"
	alistUUID := "fakeList123"
	aList := alist.NewTypeV1()
	aList.Uuid = alistUUID
	acl.CreateListRole(alistUUID)
	acl.GrantListReadAccess(userUUID, alistUUID)
	roles := acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal(1, len(roles))
	suite.True(acl.enforcer.HasRoleForUser(userUUID, "fakeList123:read"))
	suite.True(acl.enforcer.Enforce(userUUID, alistUUID, "read"))
	suite.True(acl.HasUserListReadAccess(userUUID, aList))

	// Follow the path if the user is the owner of the list
	aList.User.Uuid = userUUID
	suite.True(acl.HasUserListReadAccess(userUUID, aList))
	aList.User.Uuid = ""

	acl.RevokeListReadAccess(userUUID, alistUUID)
	roles = acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal(0, len(roles))
	suite.False(acl.enforcer.HasRoleForUser(userUUID, "fakeList123:read"))
	suite.False(acl.enforcer.Enforce(userUUID, alistUUID, "read"))
	suite.False(acl.HasUserListReadAccess(userUUID, aList))
}

func (suite *AclSuite) TestDeleteRoleWithGrantSet() {
	userUUID := "fakeUser123"
	alistUUID := "fakeList123"
	aList := alist.NewTypeV1()
	aList.Uuid = alistUUID
	acl.CreateListRole(alistUUID)
	acl.GrantListReadAccess(userUUID, alistUUID)
	suite.Equal(3, len(acl.enforcer.GetAllSubjects()))
	suite.Equal(1, len(acl.enforcer.GetAllRoles()))
	suite.True(acl.HasUserListReadAccess(userUUID, aList))
	acl.DeleteListRole(alistUUID)
	suite.Equal(1, len(acl.enforcer.GetAllSubjects()))
	suite.Equal(0, len(acl.enforcer.GetAllRoles()))
	suite.False(acl.HasUserListReadAccess(userUUID, aList))
}
