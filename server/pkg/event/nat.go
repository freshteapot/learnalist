package event

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

type natBus struct {
	sc        stan.Conn
	sub       stan.Subscription
	listeners []interface{}
}

func NewNatBus(clusterID string, clientID string, nc *nats.Conn) *natBus {
	fmt.Println("clientID", clientID)
	fmt.Println("clusterID", clusterID)
	sc, err := stan.Connect(clusterID, clientID,
		stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			log.Fatalf("Connection lost, reason: %v", reason)
		}),
		stan.Pings(10, 5),
	)

	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, nc.Opts.Url)
	}

	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", nc.Opts.Servers, clusterID, clientID)
	return &natBus{
		sc: sc,
	}
}

func (b *natBus) GetStan() stan.Conn {
	return b.sc
}

func (b *natBus) Publish(topic string, args ...interface{}) {
	if err := b.sc.Publish(topic, args[0].([]byte)); err != nil {
		log.Fatal(err)
	}
}

func (b *natBus) Close(topic string) {
	err := b.sub.Close()
	if err != nil {
		fmt.Printf("error closing stan sub: %v\n", err)
	}

	err = b.sc.Close()
	if err != nil {
		fmt.Printf("error closing stan: %v\n", err)
	}
}

func (b *natBus) Listen(fn interface{}) {
	b.listeners = append(b.listeners, fn)
}

func (b *natBus) Subscribe(topic string, fn interface{}) error {
	if b.sub != nil {
		fmt.Println("TODO fix this")
		return nil
	}

	var err error
	/*
		fmt.Println("How many times")
		var err error
		durable := "internal-system"
		startOpt := stan.StartWithLastReceived()
		qgroup := ""
		mcb := func(msg *stan.Msg) {
			var entryLog Eventlog
			err := json.Unmarshal(msg.Data, &entryLog)
			if err != nil {
				return
			}

			type HandlerType func(entry Eventlog)
			if f, ok := fn.(func(entry Eventlog)); ok {
				HandlerType(f)(entryLog)
			}
		}

		b.sub, err = b.sc.QueueSubscribe(topic, qgroup, mcb, startOpt, stan.DurableName(durable))
	*/
	// TODO do something with fn
	mcb := func(msg *stan.Msg) {
		var entryLog Eventlog
		err := json.Unmarshal(msg.Data, &entryLog)
		if err != nil {
			return
		}

		// TODO change to channel
		for _, listener := range b.listeners {
			type HandlerType func(entry Eventlog)
			if f, ok := listener.(func(entry Eventlog)); ok {
				HandlerType(f)(entryLog)
			}
		}
	}

	durableName := "internal-system"
	b.sub, err = b.sc.Subscribe(topic, mcb, stan.DurableName(durableName))

	if err != nil {
		b.sc.Close()
		log.Fatalf("Failed to start subscription on '%s': %v", topic, err)
	}

	return nil
}

func (b *natBus) Unsubscribe(topic string, fn interface{}) error {
	return b.sub.Unsubscribe()
}
