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
EVENTS_STAN_CLUSERTID="test-cluster" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make run-api-server
```


```sh
cd server
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSERTID="test-cluster" \
EVENTS_NATS_SERVER="127.0.0.1" \
EVENTS_SLACK_WEBHOOK="XXX" \
go run main.go --config=../config/dev.config.yaml tools slack-events


HUGO_EXTERNAL=false \
make run-api-server
```


# Reference
- https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kubectl/
