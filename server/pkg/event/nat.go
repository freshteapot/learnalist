package event

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	messagebus "github.com/vardius/message-bus"
)

type natBus struct {
	sc stan.Conn
}

func NewNatBus(clusterID string, clientID string, nc *nats.Conn) messagebus.MessageBus {
	sc, err := stan.Connect(clusterID, clientID,
		stan.NatsConn(nc),
		stan.Pings(10, 5),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
	)

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
			var entryLog Eventlog
			err := json.Unmarshal(stanMsg.Data, &entryLog)
			if err != nil {
				return
			}

			type HandlerType func(entry Eventlog)
			if f, ok := fn.(func(entry Eventlog)); ok {
				HandlerType(f)(entryLog)
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
