package event

import (
	"encoding/json"

	messagebus "github.com/vardius/message-bus"
)

const (
	ApiUserDelete             = "api.user.delete"
	ApiUserLogin              = "api.user.login"
	ApiUserRegister           = "api.user.register"
	TopicMonolog              = "lal.monolog"
	KindUserRegisterUsername  = "username"
	KindUserRegisterIDPGoogle = "idp:google"
)

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

var (
	queueSize = 100
	bus       messagebus.MessageBus
)

func GetBus() messagebus.MessageBus {
	return bus
}

func SetBus(newBus messagebus.MessageBus) {
	bus = newBus
}

func NewEventLog(kind string, data interface{}) Eventlog {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	return Eventlog{
		Kind: kind,
		Data: b,
	}
}

func EventLogToBytes(e Eventlog) []byte {
	b, _ := json.Marshal(e)
	return b
}
