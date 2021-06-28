package event

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/utils"
	"github.com/nats-io/stan.go"
)

const (
	ApiAppSettingsRemindV1      = "api.appsettings.remind_v1"
	ApiPlank                    = "api.plank"
	CMDUserDelete               = "cmd.user.delete"
	SystemUserDelete            = "system.user.delete"
	ApiUserDelete               = "api.user.delete"
	ApiUserLogin                = "api.user.login"
	ApiUserLogout               = "api.user.logout"
	BrowserUserLogout           = "browser.user.logout"
	ApiUserRegister             = "api.user.register"
	ApiListSaved                = "api.list.saved"
	ApiListDelete               = "api.list.delete"
	ApiSpacedRepetition         = "api.spacedrepetition"
	SystemSpacedRepetition      = "system.spacedRepetition"
	SystemListDelete            = "system.list.delete"
	ApiSpacedRepetitionOvertime = "api.spacedrepetition.overtime"

	TopicMonolog                    = "lal.monolog"
	TopicStaticSite                 = "lal.staticSite"
	TopicNotifications              = "notifications"
	TopicPayments                   = "payments"
	KindUserRegisterUsername        = "username"
	KindUserRegisterIDPApple        = "idp:apple"
	KindUserLoginIDPApple           = "idp:apple"
	KindUserLoginIDPAppleViaIdToken = "idp:apple:idtoken"

	KindUserRegisterIDPGoogle        = "idp:google"
	KindUserLoginIDPGoogle           = "idp:google"
	KindUserLoginIDPGoogleViaIdToken = "idp:google:idtoken"
	KindUserLoginUsername            = "username"
	KindUserLogoutSession            = "logout.session"
	KindUserLogoutSessions           = "logout.sessions"
	KindPushNotification             = "push-notification"
	KindPaymentsStripe               = "payments-stripe"

	ActionNew            = "new"
	ActionCreated        = "created"
	ActionUpdated        = "updated"
	ActionDeleted        = "deleted"
	ActionUpsert         = "upsert"
	ChangesetChallenge   = "changeset.challenge"
	ChangesetAlistPublic = "changeset.alist.public"
	ChangesetAlistUser   = "changeset.alist.user"
	ChangesetAlistList   = "changeset.alist.list"
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

type NatsSubscriber interface {
	Subscribe(topic string, sc stan.Conn) error
	Close()
}

type Eventlog struct {
	UUID        string      `json:"uuid,omitempty"`
	Kind        string      `json:"kind"`
	Data        interface{} `json:"data,omitempty"`
	Timestamp   int64       `json:"timestamp,omitempty"`
	Action      string      `json:"action,omitempty"`
	TriggeredBy string      `json:"triggered_by,omitempty"`
}

type EventNewUser struct {
	UUID string      `json:"uuid"`
	Kind string      `json:"kind"`
	Data interface{} `json:"data,omitempty"`
}

type EventUser struct {
	UUID string `json:"uuid"`
	Kind string `json:"kind"`
}

type EventList struct {
	UUID     string      `json:"uuid"`
	UserUUID string      `json:"user_uuid"`
	Action   string      `json:"action,omitempty"`
	Data     alist.Alist `json:"data,omitempty"` // If the list is not present it fails on json.Unmarshal
}

type EventListOwner struct {
	UUID     string `json:"uuid"`
	UserUUID string `json:"user_uuid"`
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
