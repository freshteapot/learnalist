package acl

import (
	"fmt"

	"github.com/casbin/casbin"
	"github.com/freshteapot/learnalist-api/server/api/alist"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
	sqlxadapter "github.com/memwey/casbin-sqlx-adapter"
)

type Acl struct {
	enforcer *casbin.Enforcer
}

func NewAclFromModel(dataSourceName string) *Acl {
	// rbac_model.conf
	modelText := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

	// TODO share the same from database creation
	dataSourceName = "file:" + dataSourceName
	adapter := sqlxadapter.NewAdapter("sqlite3", dataSourceName)
	model := casbin.NewModel(modelText)
	enforcer := casbin.NewEnforcer(model, adapter)
	acl := &Acl{
		enforcer: enforcer,
	}
	enforcer.LoadPolicy()

	acl.Init()
	return acl
}

// Make sure this can be ran over and over without issue.
func (acl Acl) Init() {
	acl.createPublicRole()
}

// createPublicRole Internal method, "public:write", is the concept
// that a user, is allowed to make public lists. By default, all lists are private.
func (acl Acl) createPublicRole() {
	acl.enforcer.AddPolicy("public:write", "public", "write")
}

// CreateListRoles create roles, to allow users to have read or write access to a list.
func (acl Acl) CreateListRoles(alistUUID string) {
	read := getRoleKeyListRead(alistUUID)
	write := getRoleKeyListWrite(alistUUID)
	acl.enforcer.AddPolicy(read, alistUUID, "read")
	acl.enforcer.AddPolicy(write, alistUUID, "write")

	acl.MakeListPrivateForOwner(alistUUID)
}

func (acl Acl) DeleteListRoles(alistUUID string) {
	read := getRoleKeyListRead(alistUUID)
	write := getRoleKeyListWrite(alistUUID)
	share := getRoleKeyListShare(alistUUID)
	// Remove the policy
	acl.enforcer.RemovePolicy(read, alistUUID, "read")
	acl.enforcer.RemovePolicy(write, alistUUID, "write")

	// Remove access to the deleted policy
	acl.enforcer.RemoveFilteredGroupingPolicy(1, read)
	acl.enforcer.RemoveFilteredGroupingPolicy(1, write)
	acl.enforcer.RemoveFilteredPolicy(0, share)

}

// GrantListPublicWriteAccess will allow the user to publish lists to the public section.
// By default all lists are private.
func (acl Acl) GrantListPublicWriteAccess(userUUID string) {
	acl.enforcer.AddRoleForUser(userUUID, "public:write")
}

// RevokeListPublicWriteAccess remove a users access to creating public lists.
func (acl Acl) RevokeListPublicWriteAccess(userUUID string) {
	acl.enforcer.DeleteRoleForUser(userUUID, "public:write")
}

// GrantListReadAccess grant access to the user to be able to read the list.
func (acl Acl) GrantListReadAccess(userUUID string, alistUUID string) {
	// TODO should I check shared access?
	read := getRoleKeyListRead(alistUUID)
	acl.enforcer.AddRoleForUser(userUUID, read)
}

func (acl Acl) RevokeListReadAccess(userUUID string, alistUUID string) {
	read := getRoleKeyListRead(alistUUID)
	acl.enforcer.DeleteRoleForUser(userUUID, read)
}

func (acl Acl) HasUserListReadAccess(userUUID string, aList *alist.Alist) bool {
	if userUUID == aList.User.Uuid {
		return true
	}
	return acl.enforcer.Enforce(userUUID, aList.Uuid, "read")
}

func (acl Acl) HasUserPublicWriteAccess(userUUID string) bool {
	return acl.enforcer.Enforce(userUUID, "public", "write")
}

func (acl Acl) MakeListPublic(alistUUID string) {
	share := getRoleKeyListShare(alistUUID)
	acl.enforcer.RemoveFilteredPolicy(0, share)
	acl.enforcer.AddPolicy(share, alistUUID, "public")
}

func (acl Acl) MakeListShared(alistUUID string) {
	share := getRoleKeyListShare(alistUUID)
	acl.enforcer.RemoveFilteredPolicy(0, share)
	acl.enforcer.AddPolicy(share, alistUUID, "shared")
}

func (acl Acl) MakeListPrivateForOwner(alistUUID string) {
	// This magically removes the users, but not the actual policy
	// Much to learn
	read := getRoleKeyListRead(alistUUID)
	acl.enforcer.RemoveFilteredGroupingPolicy(1, read)

	share := getRoleKeyListShare(alistUUID)
	acl.enforcer.RemoveFilteredPolicy(0, share)
	acl.enforcer.AddPolicy(share, alistUUID, "private")
}

func (acl Acl) IsListPublic(alistUUID string) bool {
	return acl.isListShared(alistUUID, "public")
}

func (acl Acl) IsListPrivate(alistUUID string) bool {
	return acl.isListShared(alistUUID, "private")
}

func (acl Acl) IsListShared(alistUUID string) bool {
	return acl.isListShared(alistUUID, "shared")
}

func (acl Acl) isListShared(alistUUID string, shareType string) bool {
	share := getRoleKeyListShare(alistUUID)
	return acl.enforcer.HasPolicy(share, alistUUID, shareType)
}

func getRoleKeyListRead(alistUUID string) string {
	return fmt.Sprintf("%s:read", alistUUID)
}

func getRoleKeyListWrite(alistUUID string) string {
	return fmt.Sprintf("%s:write", alistUUID)
}

func getRoleKeyListShare(alistUUID string) string {
	return fmt.Sprintf("%s:list:share", alistUUID)
}
