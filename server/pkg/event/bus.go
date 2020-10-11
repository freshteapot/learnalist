package event

import (
	"encoding/json"
)

func GetBus() EventlogPubSub {
	return bus
}

func SetBus(newBus EventlogPubSub) {
	bus = newBus
}

func EventLogToBytes(e Eventlog) []byte {
	b, _ := json.Marshal(e)
	return b
}
