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
	Enforcer *casbin.Enforcer
}

func NewAclFromConfig(config string, dataSourceName string) *Acl {
	// TODO share the same from database creation
	dataSourceName = "file:" + dataSourceName
	// TODO check if it exits?
	//config = "./rbac_model.conf"
	adapter := sqlxadapter.NewAdapter("sqlite3", dataSourceName)
	e := casbin.NewEnforcer(config, adapter)
	acl := NewAcl(e)
	e.LoadPolicy()
	return acl
}

func NewAcl(enforcer *casbin.Enforcer) *Acl {
	acl := &Acl{
		Enforcer: enforcer,
	}
	return acl
}

// Make sure this can be ran over and over without issue.
func (acl Acl) Init() {
	acl.createPublicRole()
}

func (acl Acl) createPublicRole() {
	acl.Enforcer.AddPolicy("public:write", "public", "write")
}

func (acl Acl) CreateListRole(alistUUID string) {
	read := fmt.Sprintf("%s:read", alistUUID)
	write := fmt.Sprintf("%s:write", alistUUID)
	acl.Enforcer.AddPolicy(read, alistUUID, "read")
	acl.Enforcer.AddPolicy(write, alistUUID, "write")
}

// GrantListPublicWriteAccess will allow the user to publish lists to the public section.
// By default all lists are private.
func (acl Acl) GrantListPublicWriteAccess(userUUID string) {
	acl.Enforcer.AddRoleForUser(userUUID, "public:write")
}

// GrantListReadAccess grant access to the user to be able to read the list.
func (acl Acl) GrantListReadAccess(userUUID string, alistUUID string) {
	// TODO should I always try and create the roles?
	// acl.createListRole(alistUUID)
	read := fmt.Sprintf("%s:read", alistUUID)
	acl.Enforcer.AddRoleForUser(userUUID, read)
}

func (acl Acl) RevokeListReadAccess(userUUID string, alistUUID string) {
	// TODO should I always try and create the roles?
	// acl.createListRole(alistUUID)
	read := fmt.Sprintf("%s:read", alistUUID)
	acl.Enforcer.DeleteRoleForUser(userUUID, read)
}

func (acl Acl) HasUserListReadAccess(userUUID string, aList *alist.Alist) bool {
	fmt.Println(fmt.Sprintf("check access to list %s by user %s", aList.Uuid, userUUID))
	if userUUID == aList.User.Uuid {
		return true
	}
	return acl.Enforcer.Enforce(userUUID, aList.Uuid, "read")
}

func (acl Acl) HasUserPublicWriteAccess(userUUID string) bool {
	return acl.Enforcer.Enforce(userUUID, "public", "write")
}
