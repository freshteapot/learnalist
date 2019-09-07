package acl

import (
	"fmt"

	"github.com/casbin/casbin/v2"
	casbinModel "github.com/casbin/casbin/v2/model"
	sqlxadapter "github.com/memwey/casbin-sqlx-adapter"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
)

type Acl struct {
	enforcer *casbin.Enforcer
}

func NewAclFromModel(db *sqlx.DB) *Acl {
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

	adapter := sqlxadapter.NewAdapterByDB(db)

	model, _ := casbinModel.NewModelFromString(modelText)
	enforcer, _ := casbin.NewEnforcer(model, adapter)
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
func (acl Acl) CreateListRoles(alistUUID string, userUUID string) {
	read := getRoleKeyListRead(alistUUID)
	write := getRoleKeyListWrite(alistUUID)
	owner := getRoleKeyListOwner(alistUUID)

	acl.enforcer.AddPolicy(read, alistUUID, "read")
	acl.enforcer.AddPolicy(write, alistUUID, "write")
	acl.enforcer.AddPolicy(owner, alistUUID, "owner")

	acl.MakeListPrivateForOwner(alistUUID)

	acl.enforcer.AddRoleForUser(userUUID, owner)
}

func (acl Acl) DeleteListRoles(alistUUID string) {
	read := getRoleKeyListRead(alistUUID)
	write := getRoleKeyListWrite(alistUUID)
	share := getRoleKeyListShare(alistUUID)
	owner := getRoleKeyListOwner(alistUUID)

	// Remove the policy
	acl.enforcer.RemovePolicy(read, alistUUID, "read")
	acl.enforcer.RemovePolicy(write, alistUUID, "write")
	acl.enforcer.RemovePolicy(owner, alistUUID, "owner")

	// Remove access to the deleted policy
	acl.enforcer.RemoveFilteredGroupingPolicy(1, read)
	acl.enforcer.RemoveFilteredGroupingPolicy(1, write)
	acl.enforcer.RemoveFilteredGroupingPolicy(1, owner)
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

func (acl Acl) HasUserListReadAccess(userUUID string, alistUUID string) bool {
	var pass bool

	pass, _ = acl.enforcer.Enforce(userUUID, alistUUID, "owner")
	if pass {
		return true
	}

	if acl.IsListPublic(alistUUID) {
		return true
	}

	pass, _ = acl.enforcer.Enforce(userUUID, alistUUID, "read")
	if pass {
		return true
	}

	return false
}

func (acl Acl) HasUserPublicWriteAccess(userUUID string) bool {
	pass, _ := acl.enforcer.Enforce(userUUID, "public", "write")
	return pass
}

// MakeListPublic Make the list readable by all
func (acl Acl) MakeListPublic(alistUUID string) {
	share := getRoleKeyListShare(alistUUID)
	acl.enforcer.RemoveFilteredPolicy(0, share)
	acl.enforcer.AddPolicy(share, alistUUID, "public")
}

// MakeListPrivate Make the list private for people with read access
func (acl Acl) MakeListPrivate(alistUUID string) {
	share := getRoleKeyListShare(alistUUID)
	acl.enforcer.RemoveFilteredPolicy(0, share)
	acl.enforcer.AddPolicy(share, alistUUID, "private")
}

// MakeListPrivateForOwner Make the list private for the owner to read
func (acl Acl) MakeListPrivateForOwner(alistUUID string) {
	// This magically removes the users, but not the actual policy
	// Much to learn
	read := getRoleKeyListRead(alistUUID)
	acl.enforcer.RemoveFilteredGroupingPolicy(1, read)

	share := getRoleKeyListShare(alistUUID)
	acl.enforcer.RemoveFilteredPolicy(0, share)
	acl.enforcer.AddPolicy(share, alistUUID, "owner")
}

func (acl Acl) IsListPublic(alistUUID string) bool {
	return acl.isListShared(alistUUID, "public")
}

func (acl Acl) IsListPrivate(alistUUID string) bool {
	return acl.isListShared(alistUUID, "owner")
}

func (acl Acl) IsListShared(alistUUID string) bool {
	return acl.isListShared(alistUUID, "private")
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

func getRoleKeyListOwner(alistUUID string) string {
	return fmt.Sprintf("%s:list:owner", alistUUID)
}
