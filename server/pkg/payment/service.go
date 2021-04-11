package payment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/thoas/go-funk"

	"github.com/stripe/stripe-go/v71/webhook"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/checkout/session"
)

type SupportV1Input struct {
	PriceID string `json:"price_id"`
}

// Thank you https://lornajane.net/posts/2020/accessing-nested-config-with-viper
type PaymentOption struct {
	ID          string `json:"id"`
	Currency    string `json:"currency"`
	UnitAmount  int    `json:"unit_amount" mapstructure:"unit_amount"`
	HumanAmount int    `json:"human_amount" mapstructure:"human_amount"`
}

type createCheckoutSessionResponse struct {
	SessionID string `json:"id"`
}

type PaymentServiceConfig struct {
	Server        string
	WebhookSecret string
	PrivateKey    string
	Options       []PaymentOption
}

type PaymentService struct {
	urlPrefix  string
	config     PaymentServiceConfig
	logContext logrus.FieldLogger
}

func NewService(config PaymentServiceConfig, log logrus.FieldLogger) PaymentService {
	// This is your real test secret API key.
	stripe.Key = config.PrivateKey
	return PaymentService{
		urlPrefix:  "/",
		config:     config,
		logContext: log,
	}
}

func (s PaymentService) Serve(prefix string, router *echo.Group) error {
	// TODO not sure I need this
	s.urlPrefix = strings.TrimLeft(prefix, "/")
	// I cant use the same prefix today with echo + static site
	router.POST("/create-checkout-session", s.CreateCheckoutSession)
	router.POST("/webhook", s.Webhook)
	return nil
}

func (s PaymentService) Webhook(c echo.Context) error {
	r := c.Request()
	w := c.Response().Writer

	if r.Method != "POST" {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return nil
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("ioutil.ReadAll: %v", err)
		return nil
	}

	moment, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), s.config.WebhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return nil
	}

	event.GetBus().Publish(event.TopicPayments, event.Eventlog{
		Kind: event.KindPaymentsStripe,
		Data: moment, // I wonder if I want the raw or the stripe event?
	})

	if moment.Type == "checkout.session.completed" {
		// TODO trigger event
		// TODO get the client reference ID
		fmt.Println("Checkout Session completed!")
	}

	return c.NoContent(http.StatusOK)
}

func (s PaymentService) CreateCheckoutSession(c echo.Context) error {
	loggedInUser := c.Get("loggedInUser").(uuid.User)
	userUUID := loggedInUser.Uuid
	s.logContext.WithFields(logrus.Fields{
		"user_uuid": userUUID,
	}).Info("Make the money")

	clientReferenceID := userUUID

	var input SupportV1Input
	defer c.Request().Body.Close()
	jsonBytes, _ := ioutil.ReadAll(c.Request().Body)

	err := json.Unmarshal(jsonBytes, &input)
	if err != nil {
		response := api.HTTPResponseMessage{
			Message: "TODO: json",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	priceReferenceID := input.PriceID

	valid := funk.Contains(s.config.Options, func(option PaymentOption) bool {
		return option.ID == priceReferenceID
	})

	if !valid {
		response := api.HTTPResponseMessage{
			Message: "TODO: price_id",
		}
		return c.JSON(http.StatusBadRequest, response)
	}

	// Paths to the actual HTML
	prefix := fmt.Sprintf("%s/payment", s.config.Server)
	// This maybe should go to /payment/support.html#/action=success
	successURL := prefix + "/success.html"
	// This maybe should go to /payment/support.html#/action=cancel
	cancelURL := prefix + "/cancel.html"

	product := &stripe.CheckoutSessionLineItemParams{
		Price:    stripe.String(priceReferenceID),
		Quantity: stripe.Int64(1),
	}

	params := &stripe.CheckoutSessionParams{
		ClientReferenceID: stripe.String(clientReferenceID),
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			product,
		},

		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
	}

	session, err := session.New(params)

	if err != nil {
		// TODO need to do more than log
		log.Printf("session.New: %v", err)
	}

	data := createCheckoutSessionResponse{
		SessionID: session.ID,
	}

	return c.JSON(http.StatusOK, data)
}

// LoadOptions given a path, try to load the file and return payment options
func LoadOptions(pathTo string) ([]PaymentOption, error) {
	var options []PaymentOption
	data, err := ioutil.ReadFile(pathTo)
	if err != nil {
		return options, err
	}

	err = json.Unmarshal(data, &options)
	if err != nil {
		return options, err
	}
	return options, err
}
