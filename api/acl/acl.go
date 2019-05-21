package acl

import (
	"fmt"

	"github.com/casbin/casbin"
	"github.com/freshteapot/learnalist-api/api/alist"
	_ "github.com/mattn/go-sqlite3" // All the cool kids are doing it.
	sqlxadapter "github.com/memwey/casbin-sqlx-adapter"
)

type Acl struct {
	// TODO do I want to keep this inside only
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

func (acl Acl) createPublicRole() {
	acl.enforcer.AddPolicy("public:write", "public", "write")
}

func (acl Acl) CreateListRole(alistUUID string) {
	read := fmt.Sprintf("%s:read", alistUUID)
	write := fmt.Sprintf("%s:write", alistUUID)
	acl.enforcer.AddPolicy(read, alistUUID, "read")
	acl.enforcer.AddPolicy(write, alistUUID, "write")
}

// GrantListPublicWriteAccess will allow the user to publish lists to the public section.
// By default all lists are private.
func (acl Acl) GrantListPublicWriteAccess(userUUID string) {
	acl.enforcer.AddRoleForUser(userUUID, "public:write")
}

func (acl Acl) RevokeListPublicWriteAccess(userUUID string) {
	acl.enforcer.DeleteRoleForUser(userUUID, "public:write")
}

// GrantListReadAccess grant access to the user to be able to read the list.
func (acl Acl) GrantListReadAccess(userUUID string, alistUUID string) {
	// TODO should I always try and create the roles?
	// acl.createListRole(alistUUID)
	read := fmt.Sprintf("%s:read", alistUUID)
	acl.enforcer.AddRoleForUser(userUUID, read)
}

func (acl Acl) RevokeListReadAccess(userUUID string, alistUUID string) {
	// TODO should I always try and create the roles?
	// acl.createListRole(alistUUID)
	read := fmt.Sprintf("%s:read", alistUUID)
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