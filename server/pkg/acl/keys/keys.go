package keys

var (
	ListWriteAccessForUser = "list:%s:write:%s"
	ListReadAccessForUser  = "list:%s:read:%s"
	ListOwnerAccessForUser = "list:%s:owner:%s"
	ListSharePublic        = "list:%s:share:public"
	ListSharePrivate       = "list:%s:share:private"
	ListShareFriends       = "list:%s:share:friends"
)

const (
	SharedWithPublic  = "public"
	NotShared         = "private"
	SharedWithFriends = "friends"
)
