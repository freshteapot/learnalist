package keys

const (
	ListWriteAccessForUser = "api:list:%s:write:%s"
	ListReadAccessForUser  = "api:list:%s:read:%s"
	ListOwnerAccessForUser = "api:list:%s:owner:%s"
	ListSharePublic        = "api:list:%s:share:public"
	ListSharePrivate       = "api:list:%s:share:private"
	ListShareFriends       = "api:list:%s:share:friends"
	FilterListShare        = "api:list:%s:share:%%"
	SharedWithPublic       = "public"
	NotShared              = "private"
	SharedWithFriends      = "friends"
	ActionRevoke           = "revoke"
	ActionGrant            = "grant"
)
