package payment

import (
	"encoding/json"

	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v71"
)

type managerListener struct {
	subscription stan.Subscription
	logContext   logrus.FieldLogger
}

func NewManagerListener(
	logContext logrus.FieldLogger,
) *managerListener {
	return &managerListener{
		logContext: logContext,
	}
}

func (l *managerListener) Subscribe(topic string, sc stan.Conn) (err error) {

	handle := func(msg *stan.Msg) {
		var moment event.Eventlog
		json.Unmarshal(msg.Data, &moment)

		l.OnEvent(moment)

	}

	durableName := "payments.manager"
	l.subscription, err = sc.Subscribe(
		topic,
		handle,
		stan.DurableName(durableName),
		stan.DeliverAllAvailable(),
		stan.MaxInflight(1),
	)
	if err == nil {
		l.logContext.Info("Running")
	}
	return err
}

func (l *managerListener) Close() {
	err := l.subscription.Close()
	if err != nil {
		l.logContext.WithField("error", err).Error("closing subscription")
	}
}

func (l *managerListener) OnEvent(entry event.Eventlog) {
	// Data = stripe payload
	// eventID + type
	stripeEvent := stripe.Event{}
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &stripeEvent)
	// TODO Write to db
	// TODO need repo
	// TODO do I want this to be a different db?

	if stripeEvent.Type != "checkout.session.completed" {
		return
	}

	// TODO how to handle future payments? ie when we want todo more than just give them public access?
	// Trigger all the things
	eventID := stripeEvent.ID
	paymentIntentUUID := stripeEvent.GetObjectValue("payment_intent")
	userUUID := stripeEvent.GetObjectValue("client_reference_id")
	accessType := "grant"

	logContext := l.logContext.WithFields(logrus.Fields{
		"event_id":       eventID,
		"payment_intent": paymentIntentUUID,
		"userUUID":       userUUID,
	})

	if paymentIntentUUID == "" || userUUID == "" {
		logContext.Error("Missing data")
		return
	}

	event.GetBus().Publish(event.TopicMonolog, event.Eventlog{
		UUID: eventID,
		Kind: acl.EventPublicListAccess,
		Data: acl.EventPublicListAccessData{
			UserUUID: userUUID,
			Action:   accessType,
		},
		TriggeredBy: "payment",
	})
	logContext.Info("Paid")
}
