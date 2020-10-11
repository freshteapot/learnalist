package event

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

type natsBus struct {
	sc        stan.Conn
	sub       stan.Subscription
	listeners []eventlogPubSubListener
}

func NewNatsBus(clusterID string, clientID string, nc *nats.Conn, log logrus.FieldLogger) *natsBus {
	logContext := log.WithFields(logrus.Fields{
		"cluster_id": clusterID,
		"client_id":  clientID,
	})

	logContext.Info("Connecting to nats server...")
	sc, err := stan.Connect(clusterID, clientID,
		stan.NatsConn(nc),
		stan.SetConnectionLostHandler(func(_ stan.Conn, reason error) {
			logContext.Fatalf("Connection lost, reason: %v", reason)
		}),
		stan.Pings(10, 5),
	)

	if err != nil {
		logContext.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, nc.Opts.Url)
	}

	logContext.Info("connected to nats server")
	return &natsBus{
		sc: sc,
	}
}

func (b *natsBus) Publish(moment Eventlog) {
	mb, _ := json.Marshal(moment)
	topic := TopicMonolog
	if err := b.sc.Publish(topic, mb); err != nil {
		log.Fatal(err)
	}
}

func (b *natsBus) Close() {
	err := b.sub.Close()
	if err != nil {
		fmt.Printf("error closing stan sub: %v\n", err)
	}

	nc := b.sc.NatsConn()
	err = b.sc.Close()
	if err != nil {
		fmt.Printf("error closing stan: %v\n", err)
	}

	nc.Close()
}

func (b *natsBus) Subscribe(key string, fn interface{}) {
	listener := eventlogPubSubListener{
		key: key,
		fn:  fn,
	}
	b.listeners = append(b.listeners, listener)
}

func (b *natsBus) Start() {
	topic := TopicMonolog
	var err error
	mcb := func(msg *stan.Msg) {
		var entryLog Eventlog
		err := json.Unmarshal(msg.Data, &entryLog)
		if err != nil {
			fmt.Println(err)
			return
		}

		for _, listener := range b.listeners {
			type HandlerType func(entry Eventlog)
			if f, ok := listener.fn.(func(entry Eventlog)); ok {
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
}

func (b *natsBus) Unsubscribe(key string) {
	for index, listener := range b.listeners {
		if listener.key == key {
			b.listeners = append(b.listeners[:index], b.listeners[index+1:]...)
			break
		}
	}
}
