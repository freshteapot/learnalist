package user

type DatastoreUsers interface {
	// User
	UserExists(userUUID string) bool
}
