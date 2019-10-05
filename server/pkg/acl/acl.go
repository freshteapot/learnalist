package acl

type Acl interface {
	AclWriter
	AclReader
}

type AclWriter interface {
	GrantListWriteAccess(alistUUID string, userUUID string) error
	RevokeListWriteAccess(alistUUID string, userUUID string) error

	GrantListReadAccess(alistUUID string, userUUID string) error
	RevokeListReadAccess(alistUUID string, userUUID string) error

	ShareListWithPublic(alistUUID string) error
	ShareListWithPrivate(alistUUID string) error
	ShareListWithFriends(alistUUID string) error
}

type AclReader interface {
	HasUserListReadAccess(alistUUID string, userUUID string) (bool, error)
	HasWriteAccess(alistUUID string, userUUID string) (bool, error)
	IsListPublic(alistUUID string) (bool, error)
	IsListPrivate(alistUUID string) (bool, error)
	IsListAvailableToFriends(alistUUID string) (bool, error)
	ListIsSharedWith(alistUUID string) (string, error)
}

/*
func fun() {

	   	`CREATE TABLE IF NOT EXISTS acl_simple (
	     alist_uuid CHARACTER(36),
	     user_uuid CHARACTER(36),
	     access CHARACTER(100) not null,
	     UNIQUE(alist_uuid, user_uuid, access)
	   );`

	var access string
	alistUUID := "alist-123"
	userUUID := "user-456"
	access = grantListReadAccess(alistUUID, userUUID)
	insert(alistUUID, userUUID, access)

	access = grantListWriteAccess(alistUUID, userUUID)
	insert(alistUUID, userUUID, access)

	access = grantListOwnerAccess(alistUUID, userUUID)
	insert(alistUUID, userUUID, access)

	access = shareListWithPublic(alistUUID)
	insert(alistUUID, userUUID, access)

	access = shareListWithPrivate(alistUUID)
	insert(alistUUID, userUUID, access)

	access = shareListWithFriends(alistUUID)
	insert(alistUUID, userUUID, access)
}

func grantListWriteAccess(alistUUID string, userUUID string) string {
	return fmt.Sprintf(KeyListWriteAccessForUser, alistUUID, userUUID)
}

func grantListReadAccess(alistUUID string, userUUID string) string {
	return fmt.Sprintf(KeyListReadAccessForUser, alistUUID, userUUID)
}

func grantListOwnerAccess(alistUUID string, userUUID string) string {
	return fmt.Sprintf(KeyListOwnerAccessForUser, alistUUID, userUUID)
}

func shareListWithPublic(alistUUID string) string {
	return fmt.Sprintf(KeyListSharePublic, alistUUID)
}

func shareListWithPrivate(alistUUID string) string {
	return fmt.Sprintf(KeyListSharePrivate, alistUUID)
}

func shareListWithFriends(alistUUID string) string {
	return fmt.Sprintf(KeyListShareFriends, alistUUID)
}

func deleteList() {

}

func insert(alistUUID string, userUUID string, access string) {
	query := `
INSERT INTO acl_simple (alist_uuid, user_uuid, access)
VALUES("%s", "%s", "%s");`
	fmt.Println(fmt.Sprintf(query, alistUUID, userUUID, access))
}

/*
INSERT INTO acl_simple (alist_uuid, user_uuid, access)
VALUES("", "", "");


INSERT INTO acl_simple (alist_uuid, user_uuid, access)
VALUES("", "", "");

list:%s:owner
list:%s:share

list:%s:%s:read, alistID, userUUID

user:%s:public:write


// Who can read it
```
list:%s:read:%s, alistID, userUUID
```

// Who can read it
```
list:%s:write:%s, alistID, userUUID
```

// Who is the list owner
```
list:%s:owner:%s, alistID, userUUID
```


// Share with public
```
list:%s:share:public
```

// Share with private
```
list:%s:share:private
```

// Share with a user
```
list:%s:share:%s, alistID, userUUID
```

userUUID, alistUUID, "owner"
*/
