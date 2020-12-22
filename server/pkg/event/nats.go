package event

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

type NatsBus struct {
	sc            stan.Conn
	subscriptions map[string]stan.Subscription
	listeners     []eventlogPubSubListener
	logContext    logrus.FieldLogger
}

func NewNatsBus(clusterID string, clientID string, nc *nats.Conn, log logrus.FieldLogger) *NatsBus {
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
	return &NatsBus{
		sc:            sc,
		subscriptions: make(map[string]stan.Subscription, 0),
		logContext:    logContext,
	}
}

func (b *NatsBus) Publish(topic string, moment Eventlog) {
	if moment.Timestamp == 0 {
		moment.Timestamp = time.Now().UTC().Unix()
	}

	mb, _ := json.Marshal(moment)

	if err := b.sc.Publish(topic, mb); err != nil {
		b.logContext.Fatal(err)
	}
}

func (b *NatsBus) Close() {
	for _, sub := range b.subscriptions {
		err := sub.Close()
		if err != nil {
			b.logContext.WithFields(logrus.Fields{
				"error": err,
			}).Error("error closing stan sub")
		}
	}

	nc := b.sc.NatsConn()
	err := b.sc.Close()
	if err != nil {
		b.logContext.WithFields(logrus.Fields{
			"error": err,
		}).Error("error closing stan")
	}
	nc.Close()
}

func (b *NatsBus) Subscribe(topic string, key string, fn interface{}) {
	listener := eventlogPubSubListener{
		topic: topic,
		key:   key,
		fn:    fn,
	}
	b.listeners = append(b.listeners, listener)
}

func (b *NatsBus) Start(topic string) {
	var err error
	mcb := func(msg *stan.Msg) {
		var entryLog Eventlog
		err := json.Unmarshal(msg.Data, &entryLog)
		if err != nil {
			return
		}

		for _, listener := range b.listeners {
			if listener.topic != topic {
				continue
			}

			type HandlerType func(entry Eventlog)
			if f, ok := listener.fn.(func(entry Eventlog)); ok {
				HandlerType(f)(entryLog)
			}
		}
	}

	// Maybe I want to change to QueueSubscribe
	durableName := "internal-system"
	sub, err := b.sc.Subscribe(topic, mcb, stan.DurableName(durableName))

	if err != nil {
		b.sc.Close()
		b.logContext.WithFields(logrus.Fields{
			"error": err,
			"topic": topic,
		}).Fatal("Failed to start subscription")
	}

	b.subscriptions[topic] = sub
}

func (b *NatsBus) Unsubscribe(topic string, key string) {
	for index, listener := range b.listeners {
		if listener.topic == topic && listener.key == key {
			b.listeners = append(b.listeners[:index], b.listeners[index+1:]...)
			break
		}
	}
}

func (b *NatsBus) Connection() stan.Conn {
	return b.sc
}
