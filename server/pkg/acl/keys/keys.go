package keys

const (
	KindOwnerAccessForUser = "api:%s:%s:owner:%s"      // kind, extUUID userUUID
	KindWriteAccessForUser = "api:%s:%s:write:%s"      // kind, extUUID userUUID
	KindReadAccessForUser  = "api:%s:%s:read:%s"       // kind, extUUID userUUID
	KindSharePublic        = "api:%s:%s:share:public"  // kind, extUUID
	KindSharePrivate       = "api:%s:%s:share:private" // kind, extUUID
	KindShareFriends       = "api:%s:%s:share:friends" // kind, extUUID
	FilterListShare        = "api:list:%s:share:%%"
	SharedWithPublic       = "public"
	NotShared              = "private"
	SharedWithFriends      = "friends"
	ActionRevoke           = "revoke"
	ActionGrant            = "grant"
)
