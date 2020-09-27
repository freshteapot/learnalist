package event

import (
	"encoding/json"

	messagebus "github.com/vardius/message-bus"
)

func GetBus() messagebus.MessageBus {
	return bus
}

func SetBus(newBus messagebus.MessageBus) {
	bus = newBus
}

func EventLogToBytes(e Eventlog) []byte {
	b, _ := json.Marshal(e)
	return b
}
