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

# Fire up the registry
- make sure container-registry (docker-registry) is running
- log onto the server and port-forward 5000

```sh
ssh $SSH_SERVER sudo kubectl port-forward deployment/container-registry 5000:5000 &
```

```sh
ssh $SSH_SERVER -L 6443:127.0.0.1:6443 -N &
```

# Setup tunnel to the registry
```sh
ssh $SSH_SERVER -L 5000:127.0.0.1:5000 -N &
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


# Install
```sh
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
ssh $SSH_SERVER -L 5000:127.0.0.1:5000 -N &
```
```sh
ssh $SSH_SERVER
sudo su -
kill -9 $(lsof -ti tcp:5000)
```

```sh
kubectl scale deployment/container-registry  --replicas=1
ssh $SSH_SERVER sudo kubectl port-forward deployment/container-registry 5000:5000 &
```

Push latest image and sync site-assets (js, css)
```sh
make push-image
make sync-site-assets
```

Make sure k8s file is uptodate
```sh
kubectl apply -f k8s/learnalist.yaml
```

Patch if only bumped latest version
```sh
PATCH=$(cat <<_EOF_
  {"spec":{"template":{"metadata":{"creationTimestamp":"$(date -u '+%FT%TZ')"}}}}
  _EOF_
)
kubectl patch deployment learnalist -p "${PATCH}"
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
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -- sh
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
