package event

import (
	"encoding/json"

	messagebus "github.com/vardius/message-bus"
)

const (
	ApiUserDelete = "api.user.delete"
	ApiUserLogin  = "api.user.login"
	TopicMonolog  = "lal.monolog"
)

type Eventlog struct {
	Kind string
	Data []byte
}

var (
	queueSize = 100
	bus       messagebus.MessageBus
)

func NewMemoryBus() messagebus.MessageBus {
	return messagebus.New(queueSize)
}

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
