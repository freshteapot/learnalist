package keys

var (
	ListWriteAccessForUser = "api:list:%s:write:%s"
	ListReadAccessForUser  = "api:list:%s:read:%s"
	ListOwnerAccessForUser = "api:list:%s:owner:%s"
	ListSharePublic        = "api:list:%s:share:public"
	ListSharePrivate       = "api:list:%s:share:private"
	ListShareFriends       = "api:list:%s:share:friends"
	FilterListShare        = "api:list:%s:share:%%"
)

const (
	SharedWithPublic  = "public"
	NotShared         = "private"
	SharedWithFriends = "friends"
)
