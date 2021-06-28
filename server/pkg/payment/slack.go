package payment

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
	"github.com/slack-go/slack"
	"github.com/stripe/stripe-go/v71"
)

type PostWebhook = func(url string, msg *slack.WebhookMessage) error

type slackListener struct {
	post         PostWebhook
	webhook      string
	subscription stan.Subscription
	logContext   logrus.FieldLogger
}

func NewSlackListener(
	post PostWebhook,
	webhook string,
	logContext logrus.FieldLogger,
) *slackListener {
	return &slackListener{
		post:       post,
		webhook:    webhook,
		logContext: logContext,
	}
}

func (l *slackListener) Subscribe(topic string, sc stan.Conn) (err error) {

	handle := func(msg *stan.Msg) {
		var moment event.Eventlog
		json.Unmarshal(msg.Data, &moment)

		l.OnEvent(moment)

	}

	durableName := "payments.slack"
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

func (l *slackListener) Close() {
	err := l.subscription.Close()
	if err != nil {
		l.logContext.WithField("error", err).Error("closing subscription")
	}
}

func (l *slackListener) OnEvent(entry event.Eventlog) {
	// Data = stripe payload
	// eventID + type
	stripeEvent := stripe.Event{}
	b, _ := json.Marshal(entry.Data)
	json.Unmarshal(b, &stripeEvent)

	var msg slack.WebhookMessage

	switch stripeEvent.Type {
	case "checkout.session.completed":
		msg.Text = fmt.Sprintf(`
:money_with_wings: :money_with_wings: :money_with_wings:
user:%s
type:%s
id:%s
livemode:%t
payment_intent:%s

> curl https://api.stripe.com/v1/events/%s -u KEY:
		`,
			stripeEvent.GetObjectValue("client_reference_id"),
			stripeEvent.Type,
			stripeEvent.ID,
			stripeEvent.Livemode,
			stripeEvent.GetObjectValue("payment_intent"),
			stripeEvent.ID)

		msg.Text = strings.TrimSpace(msg.Text)
	default:
		msg.Text = fmt.Sprintf(`
:money_with_wings: :money_with_wings: :money_with_wings:
type:%s
id:%s
livemode:%t
payment_intent:%s

> curl https://api.stripe.com/v1/events/%s -u KEY:
		`,
			stripeEvent.Type,
			stripeEvent.ID,
			stripeEvent.Livemode,
			stripeEvent.GetObjectValue("payment_intent"),
			stripeEvent.ID)

		msg.Text = strings.TrimSpace(msg.Text)
	}

	// We parse this in to make it easier to mock
	err := l.post(l.webhook, &msg)
	if err != nil {
		l.logContext.Panic(err)
	}
}
