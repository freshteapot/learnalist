package payment

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/freshteapot/learnalist-api/server/api/uuid"

	"github.com/stripe/stripe-go/v71/webhook"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/checkout/session"
)

type PaymentOption struct {
	ID          string `json:"id"`
	Currency    string `json:"currency"`
	UnitAmount  int    `json:"unit_amount"`
	HumanAmount int    `json:"human_amount"`
}

type createCheckoutSessionResponse struct {
	SessionID string `json:"id"`
}

type PaymentService struct {
	paymentOptions []PaymentOption
	domain         string
	webhookSecret  string
	logContext     logrus.FieldLogger
}

func NewService(webhookSecret string, options string, log logrus.FieldLogger) PaymentService {
	options = strings.TrimSpace(options)

	s := PaymentService{
		domain:        "http://localhost:4242",
		webhookSecret: webhookSecret,
		logContext:    log,
	}

	err := json.Unmarshal([]byte(options), &s.paymentOptions)
	if err != nil {
		log.Fatal("Payment options are wrong")
	}
	return s
}

func (s PaymentService) Serve(router *echo.Group) error {
	// TODO Do I need to serve the files or can I mix and match with hugo (should be able to)
	router.GET("/create-checkout-session", s.CreateCheckoutSession)
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

	fmt.Println(string(b))

	event, err := webhook.ConstructEvent(b, r.Header.Get("Stripe-Signature"), s.webhookSecret)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("webhook.ConstructEvent: %v", err)
		return nil
	}

	if event.Type == "checkout.session.completed" {
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
	// TODO Need to get product ID
	priceReferenceID := "price_1IXSzr2Ez2ncWz0NI8f4ffJG"

	// TODO
	prefix := fmt.Sprintf("%s/purchase", s.domain)
	successURL := prefix + "/success.html"
	cancelURL := prefix + "/cancel.html"

	// TODO
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
		log.Printf("session.New: %v", err)
	}

	data := createCheckoutSessionResponse{
		SessionID: session.ID,
	}

	js, _ := json.Marshal(data)
	return c.JSON(http.StatusOK, js)
}
