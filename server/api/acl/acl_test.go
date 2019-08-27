package acl

import (
	"fmt"
	"testing"

	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/database"
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

}

func (suite *AclSuite) SetupTest() {
	acl = NewAclFromModel(database.PathToTestSqliteDb)
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

func (suite *AclSuite) TestCreateListRoles() {
	userUUIDOwner := "owner123"
	alistUUID := "fakeList123"
	acl.CreateListRoles(alistUUID, userUUIDOwner)

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

	acl.DeleteListRoles(alistUUID)
	filteredPolicy = acl.enforcer.GetFilteredPolicy(1, alistUUID)

	suite.Equal(0, len(filteredPolicy))
}

func (suite *AclSuite) TestReadAccessForOwner() {
	userUUIDOwner := "owner123"
	alistUUID := "fakeList123"
	acl.CreateListRoles(alistUUID, userUUIDOwner)
	suite.True(acl.HasUserListReadAccess(userUUIDOwner, alistUUID))
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
	userUUIDOwner := "owner123"
	userUUID := "fakeUser123"
	alistUUID := "fakeList123"

	acl.CreateListRoles(alistUUID, userUUIDOwner)
	acl.GrantListReadAccess(userUUID, alistUUID)
	roles := acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal(1, len(roles))
	suite.True(acl.enforcer.HasRoleForUser(userUUID, "fakeList123:read"))
	suite.True(acl.enforcer.Enforce(userUUID, alistUUID, "read"))
	suite.True(acl.HasUserListReadAccess(userUUID, alistUUID))

	// Follow the path if the user is the owner of the list
	suite.True(acl.HasUserListReadAccess(userUUID, alistUUID))

	acl.RevokeListReadAccess(userUUID, alistUUID)
	roles = acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal(0, len(roles))
	suite.False(acl.enforcer.HasRoleForUser(userUUID, "fakeList123:read"))
	suite.False(acl.enforcer.Enforce(userUUID, alistUUID, "read"))
	suite.False(acl.HasUserListReadAccess(userUUID, alistUUID))
}

func (suite *AclSuite) TestDeleteRoleWithGrantSet() {
	userUUIDOwner := "owner123"
	userUUID := "fakeUser123"
	alistUUID := "fakeList123"

	acl.CreateListRoles(alistUUID, userUUIDOwner)
	acl.GrantListReadAccess(userUUID, alistUUID)
	suite.Equal(5, len(acl.enforcer.GetAllSubjects()))
	suite.Equal(2, len(acl.enforcer.GetAllRoles()))
	suite.True(acl.HasUserListReadAccess(userUUID, alistUUID))
	acl.DeleteListRoles(alistUUID)
	suite.Equal(1, len(acl.enforcer.GetAllSubjects()))
	suite.Equal(0, len(acl.enforcer.GetAllRoles()))
	suite.False(acl.HasUserListReadAccess(userUUID, alistUUID))
}

func (suite *AclSuite) TestGetAllForAUser() {
	userUUID := "fakeUser123"
	roles := acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal(len(roles), 0)
	acl.GrantListPublicWriteAccess(userUUID)
	roles = acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal(roles[0], "public:write")
}

func (suite *AclSuite) TestGelAllLists() {
	alistUUIDs := []string{
		"fake123",
		"fake345",
		"fake567",
	}
	userUUIDOwner := "owner123"
	for _, alistUUID := range alistUUIDs {
		aList := alist.NewTypeV1()
		aList.Uuid = alistUUID
		acl.CreateListRoles(alistUUID, userUUIDOwner)
	}

	// array
	items := acl.enforcer.GetFilteredPolicy(2, "read")
	uuids := make([]string, 0)
	for _, item := range items {
		uuids = append(uuids, item[1])
	}
	suite.Equal(alistUUIDs, uuids)
}

func (suite *AclSuite) TestGetAllUsersForList() {
	userUUIDOwner := "owner123"
	userUUIDs := []string{
		"fakeUser-123",
		"fakeUser-456",
		"fakeUser-789",
	}
	alistUUIDs := []string{
		"fake123",
		"fake345",
		"fake567",
	}

	for _, alistUUID := range alistUUIDs {
		aList := alist.NewTypeV1()
		aList.Uuid = alistUUID
		acl.CreateListRoles(alistUUID, userUUIDOwner)
	}

	alistUUID := alistUUIDs[1]
	read := fmt.Sprintf("%s:read", alistUUID)
	users := acl.enforcer.GetUsersForRole(read)
	suite.Equal(0, len(users))

	for _, userUUID := range userUUIDs {
		acl.GrantListReadAccess(userUUID, alistUUID)
	}
	users = acl.enforcer.GetUsersForRole(read)
	suite.Equal(3, len(users))
}

func (suite *AclSuite) TestGelAllReadListsForUser() {
	userUUIDOwner := "owner123"
	userUUID := "fakeUser123"
	alistUUIDs := []string{
		"fake123",
		"fake345",
		"fake567",
	}

	roles := acl.enforcer.GetRolesForUser(userUUID)
	suite.Equal(0, len(roles))

	for _, alistUUID := range alistUUIDs {
		aList := alist.NewTypeV1()
		aList.Uuid = alistUUID
		acl.CreateListRoles(alistUUID, userUUIDOwner)
	}

	alistUUID := alistUUIDs[0]
	read := fmt.Sprintf("%s:read", alistUUID)
	users := acl.enforcer.GetUsersForRole(read)
	suite.Equal(0, len(users))
	acl.GrantListReadAccess(userUUID, alistUUID)
	users = acl.enforcer.GetUsersForRole(read)
	suite.Equal(1, len(users))
}

func (suite *AclSuite) TestListShareAccessIsPublic() {
	userUUIDOwner := "owner123"
	userUUID := "fakeUser123"
	alistUUID := "fake123"
	aList := alist.NewTypeV1()
	aList.Uuid = alistUUID
	acl.CreateListRoles(alistUUID, userUUIDOwner)
	acl.GrantListReadAccess(userUUID, alistUUID)
	acl.MakeListPublic(alistUUID)
	suite.True(acl.IsListPublic(alistUUID))
	suite.False(acl.IsListShared(alistUUID))
	suite.False(acl.IsListPrivate(alistUUID))
}

func (suite *AclSuite) TestListShareAccessIsPrivateByDefault() {
	userUUIDOwner := "owner123"
	userUUID := "fakeUser123"
	alistUUID := "fake123"
	aList := alist.NewTypeV1()
	aList.Uuid = alistUUID
	acl.CreateListRoles(alistUUID, userUUIDOwner)
	acl.GrantListReadAccess(userUUID, alistUUID)
	suite.False(acl.IsListPublic(alistUUID))
	suite.False(acl.IsListShared(alistUUID))
	suite.True(acl.IsListPrivate(alistUUID))
}

func (suite *AclSuite) TestListShareAccessIsPrivateAfterPublic() {
	userUUIDOwner := "owner123"
	userUUID := "fakeUser123"
	alistUUID := "fake123"
	aList := alist.NewTypeV1()
	aList.Uuid = alistUUID
	acl.CreateListRoles(alistUUID, userUUIDOwner)
	acl.GrantListReadAccess(userUUID, alistUUID)
	acl.MakeListPublic(alistUUID)
	suite.True(acl.IsListPublic(alistUUID))
	suite.False(acl.IsListPrivate(alistUUID))
	acl.MakeListPrivateForOwner(alistUUID)
	suite.False(acl.IsListPublic(alistUUID))
	suite.True(acl.IsListPrivate(alistUUID))
}

func (suite *AclSuite) TestListShareAccessIsShared() {
	userUUIDOwner := "owner123"
	userUUID := "fakeUser123"
	alistUUID := "fake123"
	aList := alist.NewTypeV1()
	aList.Uuid = alistUUID
	acl.CreateListRoles(alistUUID, userUUIDOwner)
	acl.MakeListShared(alistUUID)
	acl.GrantListReadAccess(userUUID, alistUUID)
	suite.False(acl.IsListPublic(alistUUID))
	suite.True(acl.IsListShared(alistUUID))
	suite.False(acl.IsListPrivate(alistUUID))
}
