package acl

type Acl interface {
	AclWriter
	AclReader
}
type AclWriter interface {
	AclWriterList
	AclWriterAsset
	AclWriterChallenge
}

type AclList interface {
	AclWriterList
	AclReaderList
}

type AclAsset interface {
	AclReaderAsset
	AclWriterAsset
}

type AclChallenge interface {
	AclReaderChallenge
	AclWriterChallenge
}

type AclReaderChallenge interface {
	HasUserChallengeWriteAccess(extUUID string, userUUID string) (bool, error)
}

type AclWriterChallenge interface {
	GrantUserChallengeWriteAccess(extUUID string, userUUID string) error
	RevokeUserChallengeWriteAccess(extUUID string, userUUID string) error
	ShareChallengeWithPublic(extUUID string) error
	// Share an challenge only with yourself, this should remove any previous access rules
	MakeChallengePrivate(extUUID string, userUUID string) error
	DeleteChallenge(extUUID string) error // TODO rename
}

type AclReaderAsset interface {
	HasUserAssetReadAccess(extUUID string, userUUID string) (bool, error)
	IsListPublic(extUUID string) (bool, error)
	IsListPrivate(extUUID string) (bool, error)
}

type AclWriterAsset interface {
	// Grant a user access to read an asset
	GrantUserAssetReadAccess(extUUID string, userUUID string) error
	// Revoke a users access to read an asset
	RevokeUserAssetReadAccess(extUUID string, userUUID string) error
	ShareAssetWithPublic(extUUID string) error
	// Share an asset only with yourself, this should remove any previous access rules
	MakeAssetPrivate(extUUID string, userUUID string) error
	DeleteAsset(extUUID string) error // TODO rename
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

type AclReaderList interface {
	HasUserListReadAccess(alistUUID string, userUUID string) (bool, error)
	HasUserListWriteAccess(alistUUID string, userUUID string) (bool, error)
	IsListPublic(alistUUID string) (bool, error)
	IsListPrivate(alistUUID string) (bool, error)
	IsListAvailableToFriends(alistUUID string) (bool, error)
	ListIsSharedWith(alistUUID string) (string, error)
}

type AclReader interface {
	AclReaderList
	AclReaderAsset
	AclReaderChallenge
}
