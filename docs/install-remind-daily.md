# Install remind daily
## Setup volume
```sh
kubectl apply -f k8s/remind-daily-data.yaml
```


kubectl create configmap learnalist-db --from-file=./server/db/

```sh
ssh $SSH_SERVER
sudo mkdir /srv/learnalist/remind-daily
```

