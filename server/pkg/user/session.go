package user

import "time"

type UserUUID string

type UserInfo struct {
	UserUUID  UserUUID
	Challenge string
	Created   time.Time
}

type UserSession struct {
	Token     string
	UserUUID  UserUUID
	Challenge string
	Created   time.Time
}

type Session interface {
	// Create create a session with a unique challenge, send the challenge in the oauth2 flow
	Create() (string, error)
	// Activate update the challenge with the userUUID and token
	Activate(session UserSession) error
	// Get session via token
	Get(token string) (UserSession, error)

	IsChallengeValid(challenge string) (bool, error)
}

type SessionMaintenance interface {
	// RemoveAllByUserUUID remove all sessions for a user
	RemoveSessionsForUser(userUUID string) error
	// RemoveExpiredChallenges remove challenges that were never activated
	RemoveExpiredChallenges() error
}

type UserWithUsernameAndPassword interface {
	Register(username string, password string) (UserUUID, error)
	// GetUserByCredentials look up the user based on username + password
	Lookup(username string, hash string) (UserUUID, error)
}

type UserFromIDP interface {
	Register(from string, kind string, identifier string, info []byte) (UserUUID, error)
	// GetUserByCredentials look up the user based on username + password
	Lookup(from string, kind string, identifier string) (UserUUID, error)
}
