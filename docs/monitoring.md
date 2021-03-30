# Monitoring



## Read the challenges stream
### Setup a tunnel
```
kubectl port-forward svc/nats 4222:4222 &
```

### Consume challenges stream
```
cd server
TOPIC=challenges \
EVENTS_STAN_CLIENT_ID=nats-reader \
EVENTS_STAN_CLUSTER_ID=stan \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools natsutils read
```

### Consume notifications stream
```
cd server
TOPIC=notifications \
EVENTS_STAN_CLIENT_ID=nats-reader \
EVENTS_STAN_CLUSTER_ID=stan \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools natsutils read
```


# How many users with tokens + name

```
SELECT
	uuid,
IFNULL(json_extract(body, '$.display_name'), uuid) AS display_name
FROM
	user_info
WHERE
	uuid IN(SELECT DISTINCT user_uuid FROM mobile_device);
```


# Read the database
```sh
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c api -- sh
sqlite3 /srv/learnalist/server.db
```


# Poormans removal of stale tokens

```sh
ssh $SSH_SERVER -L 4222:127.0.0.1:4222 -N &
```

```sh
TOPIC=notifications \
EVENTS_STAN_CLIENT_ID=nats-reader \
EVENTS_STAN_CLUSTER_ID=stan \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools natsutils read
```

```sh
cat events.ndjson | jq -r 'select(.event=="stale") | "DELETE FROM mobile_device WHERE token=\"\(.token)\";"'
```

```sh
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c api -- sh
sqlite3 /srv/learnalist/server.db
```


# Nat & Stan

## Reload config
```sh
kubectl -it exec nats-0  -- nats-server --config /etc/nats-config/nats.conf -sl reload
```

## Get stats

```sh
kubectl -it exec $(kubectl get pods -l "app=nats" -o jsonpath="{.items[0].metadata.name}")  -- wget -qO - 'localhost:8222/varz' | jq
```

## List stores
```sh
kubectl port-forward svc/stan 8222:8222 &
curl localhost:8222/streaming/storez
```

# Query for when the cert expires
```sh
export SITE_URL="learnalist.net"
export SITE_SSL_PORT="443"
openssl s_client -connect ${SITE_URL}:${SITE_SSL_PORT} \
  -servername ${SITE_URL} 2> /dev/null |  openssl x509 -noout  -dates
```

# How to link accounts
- Assuming nothing has been added
## Current user
fc
## New user
2b

## Link new user to current user
```sql
UPDATE
    user_from_idp
SET
    user_uuid="fc"
WHERE
    user_uuid="2b"
```

# User tools
## Find user
```sh
/app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools user find "iamtest1"
```

## Delete user
```sh
/app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools user delete
```

# Manage user access to public lists
/app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools list public-access XXX --access=grant

# Reference
- https://docs.nats.io/nats-streaming-concepts/monitoring/endpoints
- https://docs.nats.io/nats-server/configuration/monitoring#monitoring-endpoints
