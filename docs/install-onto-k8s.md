# Install onto kubernetes
# Notes
- You will need to add registry.devbox to your /etc/hosts

```sh
export KUBECONFIG="/Users/tinkerbell/.k3s/lal01.learnalist.net.yaml"
export SSH_SERVER="lal01.learnalist.net"
```

- remove what is currently set
- set to default based on KUBECONFIG (above)

```sh
kubectl config unset current-context
kubectl config use-context default
```

```
127.0.0.1 registry.devbox
```

# Setup file structure
- Log onto the server
```
mkdir -p /srv/learnalist/{bin,site-cache}
cp -rf ./hugo/static/* /srv/learnalist/site-cache/
cp -r ./hugo /srv/learnalist
mkdir -p /srv/learnalist/hugo/{public,content/alist,data/alist,content/alistsbyuser,data/alistsbyuser,content/challenge,data/challenge}
ls server/db/*.sql | sort | xargs cat | sqlite3 /srv/learnalist/server.db
```

# Setup tunnel to kubectl
```sh
ssh $SSH_SERVER -L 6443:127.0.0.1:6443 -N &
```

# Add resources from k8s
## Nats + stan
- First single files
- Then update the configs

# Https out of the box
- Almost ;)
- We are using the chart from https://math-nao.github.io/certs/.
- Each domain for now has its own tls.

```sh
cd ~/git/secrets/acme-certs
helm repo add certs https://math-nao.github.io/certs/charts
helm fetch --untar certs/certs
mkdir -p output
helm template certs   --name certs   --values values.yaml   --output-dir=./output
```

# Fire up the registry
- make sure container-registry (docker-registry) is running
- Setup tunnel to the registry

# Setup tunnel to the registry
```sh
kubectl port-forward deployment/container-registry 5000:5000 &
```

# Push new image
- tag and push to the registry, goes over the local tunnel.

```sh
docker push registry.devbox:5000/learnalist:latest
```


```sh
kubectl scale deployment/container-registry  --replicas=1
```


```sh
PATCH=$(cat <<_EOF_
  {"spec":{"template":{"metadata":{"creationTimestamp":"$(date -u '+%FT%TZ')"}}}}
_EOF_
)
kubectl patch deployment learnalist -p "${PATCH}"
```

## Turn off
```
ssh $SSH_SERVER
sudo su -
kill -9 $(lsof -ti tcp:5000)
```

```
kill -9 $(sudo lsof -ti tcp:5000)
```

```
kill -9 $(lsof -ti tcp:6443)
```


```sh
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c learnalist -- sh
```

## Clear when hugo is external
- **HUGO_EXTERNAL=false**
```sh
/app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools rebuild-static-site
```





kubectl create configmap learnalist-config --from-file=config.yaml=config/lal01.yaml -o yaml --dry-run | kubectl replace -f -



```sh
rsync -avzP \
--rsync-path="sudo rsync" \
${SSH_SERVER}:/srv/learnalist/server.db prod-server.db
```

# Install only assets

```sh
make build-site-assets

export KUBECONFIG="/Users/tinkerbell/.k3s/lal01.learnalist.net.yaml"
export SSH_SERVER="lal01.learnalist.net"
kubectl config unset current-context
kubectl config use-context default
kill -9 $(lsof -ti tcp:6443)
ssh $SSH_SERVER -L 6443:127.0.0.1:6443 -N &
make sync-site-assets
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c learnalist -- /app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools rebuild-static-site
```

# Install everything
```sh
make generate-openapi-js
make generate-openapi-go
make build-site-assets
make build-image-base
make build-image

export KUBECONFIG="/Users/tinkerbell/.k3s/lal01.learnalist.net.yaml"
export SSH_SERVER="lal01.learnalist.net"
kubectl config unset current-context
kubectl config use-context default
kill -9 $(lsof -ti tcp:6443)
kill -9 $(lsof -ti tcp:5000)
ssh $SSH_SERVER -L 6443:127.0.0.1:6443 -N &
```

```sh
kubectl scale deployment/container-registry  --replicas=1
kubectl port-forward deployment/container-registry 5000:5000 &
```

Push latest image and sync site-assets (js, css)
```sh
make push-image
make sync-site-assets
```

Make sure k8s file is upto date
```sh
kubectl apply -f k8s/learnalist.yaml
kubectl apply -f k8s/slack-events.yaml
kubectl apply -f k8s/event-reader.yaml
kubectl apply -f k8s/notifications-push-notifications.yaml
kubectl apply -f k8s/remind-daily.yaml
```

Patch if only bumped latest version
```sh
PATCH=$(cat <<_EOF_
  {"spec":{"template":{"metadata":{"creationTimestamp":"$(date -u '+%FT%TZ')"}}}}
_EOF_
)
kubectl patch deployment learnalist -p "${PATCH}"
kubectl patch deployment event-reader -p "${PATCH}"
kubectl patch deployment slack-events -p "${PATCH}"
kubectl patch deployment notifications-push-notifications -p "${PATCH}"
kubectl patch deployment remind-daily -p "${PATCH}"
```

```sh
kubectl scale deployment/container-registry  --replicas=0
kill -9 $(lsof -ti tcp:6443)
kill -9 $(lsof -ti tcp:5000)
kubectl config unset current-context
unset KUBECONFIG

ssh $SSH_SERVER
sudo su -
kill -9 $(lsof -ti tcp:5000)

unset SSH_SERVER
```

Remove any tunnels
```sh
ps  | grep 'ssh ' | grep 'learnalist.net' | cut -d' ' -f1 | xargs kill -9
```


# Update db
## On local machine

```sh
make sync-db-files
```

## Via a pod
```sh
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -c learnalist -- sh
```
Update tables
```sh
cat  /srv/learnalist/db/XXX | sqlite3 /srv/learnalist/server.db
```


# Query the mount
```
- name: ls
  image: "k8s.gcr.io/busybox"
  command: ["/bin/sh", "-c"]
  args: ["ls -lah /src;sleep 100000"]

  volumeMounts:
    - name: srv-learnalist-volume
      mountPath: "/src"
```



# Get configmaps

```sh
kubectl get configmaps learnalist-config -oyaml | yq r - "data[config.yaml]" > current.yaml
```


```sh
kubectl create configmap learnalist-config --from-file=config.yaml=current.yaml -o yaml --dry-run | kubectl replace -f -
```

# Setup secrets for
## Fcm
```
kubectl create secret generic learnalist-fcm \
--from-file=fcm-credentials.json=./../secrets/fcm-credentials.json
```



# Update nats configmaps
## Nats
### Current
```sh
kubectl get configmaps nats-config -oyaml | yq r - "data[nats.conf]" > nats.conf
```
### Update
```sh
kubectl create configmap stan-config --from-file=stan.conf=./k8s/nats-stan.conf -o yaml --dry-run | kubectl replace -f -
```

## Stan
### Current
```sh
kubectl get configmaps stan-config -oyaml | yq r - "data[stan.conf]" > stan.conf
```
### Update
```sh
kubectl create configmap nats-config --from-file=nats.conf=./k8s/nats-nats.conf -o yaml --dry-run | kubectl replace -f -
```

# Reload nats
```sh
nats-server --config /etc/nats-config/nats.conf -sl reload
```


# Update secret/learnalist-server

## Create
```sh
kubectl create secret generic learnalist-server \
  --from-literal=userRegisterKey="XXX"
```

## Update
```sh
kubectl create secret generic learnalist-server \
--from-literal=userRegisterKey="XXX" \
--dry-run -o yaml |
kubectl apply -f -
```
