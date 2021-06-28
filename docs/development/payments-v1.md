# Payments V1
- [ ] Store event_id + payload + created + type
- [ ] Slack event_id + type and when
- [ ] V1 event.type == checkout.session.completed
    UserID paid X
- [ ] Payment V1 event.type == checkout.session.completed
        Get userID
        Trigger public access

# Notes

id + data.object.id
data.object.client_reference_id

data.object.payment_intent.pi_1IfUHd2Ez2ncWz0N8QToTDU0

charge.succeeded includes payment_intent + ch

payment_intent


"payment_intent.created" = data.object.id
"payment_intent.succeeded" = = data.object.id
"checkout.session.completed" = data.object.payment_intent
"charge.succeeded" = data.object.payment_intent


To start with:
- store:
    - id
    - type
    - paymentIntentID (can be null if missing)
    - created

- store data + id, + type
# Pre work
- Setup the prices
- Update dev.config with the prices data

# Setup prices
## Massage the data
- u = SECRET [quickstart docs](https://stripe.com/docs/development/quickstart)
```
curl https://api.stripe.com/v1/prices \
  -u XXX \
  -d limit=100 \
  -d product=prod_J9nBoQ7DEh6GC9 \
  -G | jq  > prices.json
```

- PUBLIC = Public key
```
cat ./prices.json| jq '{
  "key": "PUBLIC",
  "options": [
    (.data[]| {id:.id, currency:.currency,unit_amount:.unit_amount, human_amount:(.unit_amount/100)})
  ]
}' > js/src/payment/v1/stripe.json
```

# Development

## Setup a webhook
```sh
stripe listen --forward-to http://192.168.0.10:1234/payments/webhook
```

## Start server
- static-site
- js
- webhook
- Alter dev.config to include prices
```sh
PAYMENT_WEBHOOK_SECRET="XXX" STATIC_SITE_EXTERNAL=false \
make clear-site rebuild-db develop
```

## Read the events

```sh
TOPIC=payments \
EVENTS_STAN_CLIENT_ID=nats-reader \
go run main.go --config=../config/dev.config.yaml \
tools natsutils read
```


## Write events to slack
```
TOPIC=payments \
EVENTS_PAYMENTS_SLACK_WEBHOOK="XXX" \
go run main.go --config=../config/dev.config.yaml \
tools payments slack
```


# Reference
- https://stripe.com/docs/checkout/integration-builder
- https://stripe.com/docs/webhooks/test
- https://stripe.com/docs/development/quickstart
