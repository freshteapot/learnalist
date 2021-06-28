package payment

import "github.com/stripe/stripe-go/v71"

type PaymentRepository interface {
	Save(event stripe.Event) error
	Get(ID string) (stripe.Event, error)
}
