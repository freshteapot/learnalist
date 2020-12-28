package event

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/api/utils"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
)

const (
	ApiPlank                         = "api.plank"
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
	TopicNotifications               = "notifications"
	KindUserRegisterUsername         = "username"
	KindUserRegisterIDPGoogle        = "idp:google"
	KindUserLoginIDPGoogle           = "idp:google"
	KindUserLoginIDPGoogleViaIdToken = "idp:google:idtoken"
	KindUserLoginUsername            = "username"
	KindUserLogoutSession            = "logout.session"
	KindUserLogoutSessions           = "logout.sessions"
	KindPushNotification             = "push-notification"
	ActionNew                        = "new"
	ActionCreated                    = "created"
	ActionUpdated                    = "updated"
	ActionDeleted                    = "deleted"
	ActionUpsert                     = "upsert"
	ChangesetChallenge               = "changeset.challenge"
)

var (
	MobileDeviceRegistered = "mobile.registered"
	MobileDeviceRemove     = "mobile.remove"
	MobileDeviceRemoved    = "mobile.removed"
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
	UUID      string      `json:"uuid,omitempty"`
	Kind      string      `json:"kind"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp,omitempty"`
	Action    string      `json:"action,omitempty"`
}

type EventUser struct {
	UUID string `json:"uuid"`
	Kind string `json:"kind"`
}

type EventList struct {
	UUID     string      `json:"uuid"`
	UserUUID string      `json:"user_uuid"`
	Action   string      `json:"action,omitempty"`
	Data     alist.Alist `json:"data,omitempty"`
}

type EventKV struct {
	UUID string      `json:"uuid"`
	Data interface{} `json:"data"`
}

type EventPlank struct {
	Action   string        `json:"action"`
	Data     openapi.Plank `json:"data"`
	UserUUID string        `json:"user_uuid"`
}

func IsUserDeleteEvent(entry Eventlog) bool {
	allowed := []string{ApiUserDelete, CMDUserDelete}
	return utils.StringArrayContains(allowed, entry.Kind)
}
