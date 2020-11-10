package event

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
)

const (
	ApiUserDelete                    = "api.user.delete"
	ApiUserLogin                     = "api.user.login"
	ApiUserLogout                    = "api.user.logout"
	BrowserUserLogout                = "browser.user.logout"
	ApiUserRegister                  = "api.user.register"
	ApiListSaved                     = "api.list.saved"
	ApiListDelete                    = "api.list.delete"
	ApiSpacedRepetition              = "api.spacedrepetition"
	TopicMonolog                     = "lal.monolog"
	KindUserRegisterUsername         = "username"
	KindUserRegisterIDPGoogle        = "idp:google"
	KindUserLoginIDPGoogle           = "idp:google"
	KindUserLoginIDPGoogleViaIdToken = "idp:google:idtoken"
	KindUserLoginUsername            = "username"
	KindUserLogoutSession            = "logout.session"
	KindUserLogoutSessions           = "logout.sessions"
)

var (
	bus EventlogPubSub
)

type eventlogPubSubListener struct {
	key string
	fn  interface{}
}
type EventlogPubSub interface {
	Start()
	Close()
	Publish(moment Eventlog)
	Subscribe(key string, fn interface{})
	Unsubscribe(key string)
}

type Eventlog struct {
	Kind string      `json:"kind"`
	Data interface{} `json:"data"`
}

type EventUser struct {
	UUID string `json:"uuid"`
	Kind string `json:"kind"`
}

type EventList struct {
	UUID     string       `json:"uuid"`
	UserUUID string       `json:"user_uuid"`
	Action   string       `json:"action,omitempty"`
	Data     *alist.Alist `json:"data,omitempty"`
}
