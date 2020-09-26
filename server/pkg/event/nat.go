package event

import (
	"fmt"
	"log"

	"github.com/nats-io/stan.go"
	messagebus "github.com/vardius/message-bus"
)

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

type natBus struct {
	sc stan.Conn
}

func NewNatBus(server string, clientID string) messagebus.MessageBus {
	sc, err := stan.Connect(server, clientID)
	fmt.Println(err)
	return &natBus{
		sc: sc,
	}
}

func (b *natBus) Publish(topic string, args ...interface{}) {
	if err := b.sc.Publish(topic, args[0].([]byte)); err != nil {
		log.Fatal(err)
	}
}

func (b *natBus) Close(topic string) {
	b.sc.Close()
}

func (b *natBus) Subscribe(topic string, fn interface{}) error {
	durableName := "TODO"
	_, err := b.sc.Subscribe(topic,
		func(stanMsg *stan.Msg) {
			type HandlerType func(msg *stan.Msg)
			if f, ok := fn.(func(msg *stan.Msg)); ok {
				HandlerType(f)(stanMsg)
			}
		},
		stan.DurableName(durableName))
	if err != nil {
		log.Fatalf("Failed to start subscription on '%s': %v", topic, err)
	}

	return nil
}

func (b *natBus) Unsubscribe(topic string, fn interface{}) error {
	return nil
}
