package event

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	messagebus "github.com/vardius/message-bus"
)

const (
	ApiUserDelete             = "api.user.delete"
	ApiUserLogin              = "api.user.login"
	ApiUserRegister           = "api.user.register"
	ApiListSaved              = "api.list.saved"
	ApiListDelete             = "api.list.delete"
	TopicMonolog              = "lal.monolog"
	KindUserRegisterUsername  = "username"
	KindUserRegisterIDPGoogle = "idp:google"
	ActionListCreated         = "created"
	ActionListUpsert          = "updated"
	ActionListDeleted         = "deleted"
)

var (
	queueSize = 100
	bus       messagebus.MessageBus
)

// Taken from https://github.com/vardius/message-bus
// So I can create a mock
type MessageBus interface {
	// Publish publishes arguments to the given topic subscribers
	// Publish block only when the buffer of one of the subscribers is full.
	Publish(topic string, args ...interface{})
	// Close unsubscribe all handlers from given topic
	Close(topic string)
	// Subscribe subscribes to the given topic
	Subscribe(topic string, fn interface{}) error
	// Unsubscribe unsubscribe handler from the given topic
	Unsubscribe(topic string, fn interface{}) error
}

type Eventlog struct {
	Kind string `json:"kind"`
	//Data []byte `json:"data"`
	Data interface{} `json:"data"`
	// TODO maybe add when
	//When int64 / time.Time
}

type EventUserRegister struct {
	UUID string `json:"uuid"`
	Kind string `json:"kind"`
}

type EventList struct {
	UUID     string       `json:"uuid"`
	UserUUID string       `json:"user_uuid"`
	Action   string       `json:"action,omitempty"`
	Data     *alist.Alist `json:"data,omitempty"`
}
