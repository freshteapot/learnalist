package user

import "time"

type UserInfoFromIDP struct {
	UserUUID   string
	IDP        string
	Identifier string
	Kind       string
	Info       string
	Created    time.Time
}

type UserInfo struct {
	UserUUID  string
	Challenge string
	Created   time.Time
}

type UserSession struct {
	Token     string
	UserUUID  string
	Challenge string
	Created   time.Time
}

type Session interface {
	// Create create a session with a unique challenge, send the challenge in the oauth2 flow
	// The string returned is the actual challenge
	Create() (string, error)
	// Activate update the challenge with the userUUID and token
	Activate(session UserSession) error
	// Get session via token
	Get(token string) (UserSession, error)

	IsChallengeValid(challenge string) (bool, error)
}

type SessionMaintenance interface {
	// RemoveSessionsForUser remove all sessions for a user
	RemoveSessionsForUser(userUUID string) error
	// RemoveExpiredChallenges remove challenges that were never activated
	RemoveExpiredChallenges() error
}

// TODO
type UserWithUsernameAndPassword interface {
	Register(username string, password string) (userUUID string, err error)
	// GetUserByCredentials look up the user based on username + password
	Lookup(username string, hash string) (userUUID string, err error)
}

type UserFromIDP interface {
	Register(idp string, identifier string, info []byte) (userUUID string, err error)
	Lookup(idp string, identifier string) (userUUID string, err error)
	GetByUserUUID(userUUID string) (UserInfoFromIDP, error)
}
