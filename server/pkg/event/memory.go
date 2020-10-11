package event

import (
	"encoding/json"

	messagebus "github.com/vardius/message-bus"
)

type lalMessageBus struct {
	bus       messagebus.MessageBus
	listeners []interface{}
}

// TODO maybe skip the whole messageBus and just use nats?
func NewMemoryBus() MessageBusWithListeners {
	return &lalMessageBus{
		bus: messagebus.New(queueSize),
	}
}

func (b *lalMessageBus) Publish(topic string, args ...interface{}) {
	var eventLog Eventlog
	err := json.Unmarshal(args[0].([]byte), &eventLog)
	if err != nil {
		panic(err)
	}
	b.bus.Publish(topic, eventLog)
}

func (b *lalMessageBus) Close(topic string) {
	b.bus.Close(topic)
}

func (b *lalMessageBus) Subscribe(topic string, fn interface{}) error {
	// TODO add listeners
	return b.bus.Subscribe(topic, fn)
}

func (b *lalMessageBus) Unsubscribe(topic string, fn interface{}) error {
	return b.bus.Unsubscribe(topic, fn)
}

func (b *lalMessageBus) Listen(fn interface{}) {
	b.listeners = append(b.listeners, fn)
}
