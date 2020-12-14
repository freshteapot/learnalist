# Monitoring



## Read the challenges stream
### Setup a tunnel
```
ssh $SSH_SERVER -L 4222:127.0.0.1:4222 -N &
ssh $SSH_SERVER sudo kubectl port-forward deployment/stan01 4222:4222 &
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
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c learnalist -- sh
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
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c learnalist -- sh
sqlite3 /srv/learnalist/server.db
```


# Get stats from nats / stan

```sh
kubectl -it exec $(kubectl get pods -l "app=nats" -o jsonpath="{.items[0].metadata.name}")  -- wget -qO - 'localhost:8222/varz' | jq
```

# Query for when the cert expires
```sh
export SITE_URL="learnalist.net"
export SITE_SSL_PORT="443"
openssl s_client -connect ${SITE_URL}:${SITE_SSL_PORT} \
  -servername ${SITE_URL} 2> /dev/null |  openssl x509 -noout  -dates
```
