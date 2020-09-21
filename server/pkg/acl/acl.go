package acl

type Acl interface {
	AclWriter
	AclReader
}
type AclWriter interface {
	AclWriterList
}

type AclWriterList interface {
	// Grant a user access to write to a list
	GrantUserListWriteAccess(alistUUID string, userUUID string) error
	// Revoke access for a user to write to a list
	RevokeUserListWriteAccess(alistUUID string, userUUID string) error

	// Grant a user access to read a list
	GrantUserListReadAccess(alistUUID string, userUUID string) error
	// Revoke a users access to read a list
	RevokeUserListReadAccess(alistUUID string, userUUID string) error

	// Share a list with the public
	ShareListWithPublic(alistUUID string) error
	// Share a list only with yourself, this should remove any previous access rules
	MakeListPrivate(alistUUID string, userUUID string) error
	// Share with friends
	ShareListWithFriends(alistUUID string) error

	DeleteList(alistUUID string) error    // TODO rename
	DeleteByExtUUID(extUUID string) error // TODO rename
}

type AclReaderKind interface {
	HasUserKindReadAccess(kind string, extUUID string, userUUID string) (bool, error)
	IsKindPublic(kind string, extUUID string) (bool, error)
	IsKindPrivate(kind string, extUUID string) (bool, error)
	IsKindAvailableToFriends(kind string, extUUID string) (bool, error)
	//HasUserKindReadAccess(alistUUID string, userUUID string) (bool, error)
}

type AclReaderList interface {
	HasUserListReadAccess(alistUUID string, userUUID string) (bool, error)
	HasUserListWriteAccess(alistUUID string, userUUID string) (bool, error)
	IsListPublic(alistUUID string) (bool, error)
	IsListPrivate(alistUUID string) (bool, error)
	IsListAvailableToFriends(alistUUID string) (bool, error)
	ListIsSharedWith(alistUUID string) (string, error)
}

type AclReader interface {
	AclReaderKind
	AclReaderList
}
