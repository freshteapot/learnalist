```sh
kubectl create secret generic slack \
  --from-literal=webhook_learnalist_events=XXX
```

```sh
kubectl get secrets slack -ojson | jq -r '.data["webhook_learnalist_events"]' | base64 -d -
```

```sh
make clear-site rebuild-db
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make run-api-server
```


```sh
cd server
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_NATS_SERVER="127.0.0.1" \
EVENTS_SLACK_WEBHOOK="XXX" \
go run main.go --config=../config/dev.config.yaml tools slack-events


HUGO_EXTERNAL=false \
make run-api-server
```


 EVENTS_STAN_CLUSTERID="test-cluster" \
EVENTS_STAN_CLIENT_ID="",

```sh
make clear-site rebuild-db
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-server" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make run-api-server
```

```
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-event-reader" \
EVENTS_NATS_SERVER="127.0.0.1" \
go run main.go --config=../config/dev.config.yaml tools event-reader
```


# Reference
- https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kubectl/




# Issues
- https://github.com/nats-io/nats-streaming-server/issues/759


```
cd ~/git/stan.go
go run examples/stan-sub/main.go -c test-cluster  -id lal-server  -durable internal-system  -unsubscribe lal.monolog
```

```sh
[1] 2020/10/10 20:30:54.279227 [DBG] STREAM: [Client:lal-server] Removed durable subscription, subject=lal.monolog, inbox=_INBOX.krPgHZCDqDkqTNibYsAzar, durable=internal-system, subid=8
[1] 2020/10/10 20:30:54.284182 [DBG] STREAM: [Client:lal-server] Closed (Inbox=_INBOX.krPgHZCDqDkqTNibYsAzUj)
[1] 2020/10/10 20:30:54.286469 [DBG] 172.17.0.1:36458 - cid:5 - Client connection closed: Client Closed
```


2020/10/10 22:31:39 Failed to start subscription on 'lal.monolog': stan: duplicate durable registration

```sh
Started new durable subscription, subject=lal.monolog, inbox=_INBOX.krPgHZCDqDkqTNibYsAzar, durable=internal-system, subid=8, sending from beginning, seq=67
```

Fails
```sh
[1] 2020/10/10 20:31:39.283431 [DBG] STREAM: [Client:lal-server] Started new durable subscription, subject=lal.monolog, inbox=_INBOX.cwA0xVciYUfZLBSIZrYSCU, durable=internal-system, subid=9, sending new-only, seq=67
[1] 2020/10/10 20:31:39.284529 [ERR] STREAM: [Client:lal-server] Duplicate durable subscription registration
[1] 2020/10/10 20:31:39.286565 [DBG] 172.17.0.1:36462 - cid:6 - Client connection closed: Client Closed
```


```
[1] 2020/10/10 20:41:30.022119 [DBG] STREAM: [Client:lal-server] Suspended durable subscription, subject=lal.monolog, inbox=_INBOX.QOPLdUzDOOTUjsdmnAMYJJ, durable=internal-system, subid=9
[1] 2020/10/10 20:41:30.023676 [DBG] STREAM: [Client:lal-server] Closed (Inbox=_INBOX.QOPLdUzDOOTUjsdmnAMY5J)
[1] 2020/10/10 20:41:30.024548 [DBG] STREAM: [Client:lal-server] Replaced old client (Inbox=_INBOX.KqQ01cgIystKyJXyg31wXB)
[1] 2020/10/10 20:41:30.029130 [DBG] STREAM: [Client:lal-server] Resumed durable subscription, subject=lal.monolog, inbox=_INBOX.KqQ01cgIystKyJXyg31wr3, durable=internal-system, subid=9
[1] 2020/10/10 20:41:30.030861 [ERR] STREAM: [Client:lal-server] Duplicate durable subscription registration
```




```
docker run \
-p 4222:4222 \
-p 8222:8222 \
-v /tmp/nats-store/:/tmp/nats-store/ nats-streaming:alpine3.12 \
--max_age 10s \
--store=FILE \
--dir=/tmp/nats-store \
--file_auto_sync=1ms \
--stan_debug=true \
--debug=true \
--http_port 8222
```
