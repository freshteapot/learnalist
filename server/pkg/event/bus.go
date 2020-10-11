package event

import (
	"encoding/json"
)

func GetBus() MessageBusWithListeners {
	return bus
}

func SetBus(newBus MessageBusWithListeners) {
	bus = newBus
}

func EventLogToBytes(e Eventlog) []byte {
	b, _ := json.Marshal(e)
	return b
}
