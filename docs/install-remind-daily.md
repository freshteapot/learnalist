# Install remind manager

## Setup volume
```sh
kubectl apply -f k8s/remind-daily-data.yaml
```


kubectl create configmap learnalist-db --from-file=./server/db/

```sh
ssh $SSH_SERVER
sudo mkdir /srv/learnalist/remind-daily
```



## Update
```sh
kubectl create configmap learnalist-db --from-file=./server/db/ -o yaml --dry-run | kubectl replace -f -
```


cat /srv/db-schema/* | sqlite3 /srv/remind-daily/remind-daily.db
cat /srv/db-schema/* | sqlite3 /srv/learnalist/server.db
