package event

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
)

const (
	CMDUserDelete                    = "cmd.user.delete"
	ApiUserDelete                    = "api.user.delete"
	ApiUserLogin                     = "api.user.login"
	ApiUserLogout                    = "api.user.logout"
	BrowserUserLogout                = "browser.user.logout"
	ApiUserRegister                  = "api.user.register"
	ApiListSaved                     = "api.list.saved"
	ApiListDelete                    = "api.list.delete"
	ApiSpacedRepetition              = "api.spacedrepetition"
	TopicMonolog                     = "lal.monolog"
	TopicStaticSite                  = "lal.staticSite"
	KindUserRegisterUsername         = "username"
	KindUserRegisterIDPGoogle        = "idp:google"
	KindUserLoginIDPGoogle           = "idp:google"
	KindUserLoginIDPGoogleViaIdToken = "idp:google:idtoken"
	KindUserLoginUsername            = "username"
	KindUserLogoutSession            = "logout.session"
	KindUserLogoutSessions           = "logout.sessions"
	KindPushNotification             = "push-notification"
	ActionCreated                    = "created"
	ActionUpdated                    = "updated"
	ActionDeleted                    = "deleted"
	ActionUpsert                     = "upsert"
	ChangesetChallenge               = "changeset.challenge"
)

var (
	bus EventlogPubSub
)

type eventlogPubSubListener struct {
	topic string
	key   string
	fn    interface{}
}
type EventlogPubSub interface {
	Start(topic string)
	Close()
	Publish(topic string, moment Eventlog)
	Subscribe(topic string, key string, fn interface{})
	Unsubscribe(topic string, key string)
}

type Eventlog struct {
	Kind      string      `json:"kind"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp,omitempty"`
	Action    string      `json:"action,omitempty"`
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

type EventKV struct {
	UUID string      `json:"uuid"`
	Data interface{} `json:"data"`
}
