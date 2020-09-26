package event

import (
	messagebus "github.com/vardius/message-bus"
)

const (
	ApiUserDelete = "api.user.delete"
)

var (
	queueSize = 100
	bus       messagebus.MessageBus
)

func init() {
	bus = messagebus.New(queueSize)
}

func GetBus() messagebus.MessageBus {
	return bus
}
