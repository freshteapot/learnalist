# Events
## Setup secrets

```sh
kubectl create secret generic slack \
  --from-literal=webhook_learnalist_events=XXX
```

## Get slack secret from the cluster
```sh
kubectl get secrets slack -ojson | jq -r '.data["webhook_learnalist_events"]' | base64 -d -
```



# Reference
- https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-kubectl/




# Issues
- https://github.com/nats-io/nats-streaming-server/issues/759


```
cd ~/git/stan.go
go run examples/stan-sub/main.go -c test-cluster  -id lal-server  -durable internal-system  -unsubscribe lal.monolog
```

