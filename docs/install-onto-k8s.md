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
kubectl exec -it $(kubectl get pods -l "app=learnalist" -o jsonpath="{.items[0].metadata.name}") -- sh
HUGO_EXTERNAL=false  /app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools rebuild-static-site
```





kubectl create configmap learnalist-config --from-file=config.yaml=config/lal01.yaml -o yaml --dry-run | kubectl replace -f -




```sh
rsync -avzP \
--rsync-path="sudo rsync" \
${SSH_SERVER}:/srv/learnalist/server.db prod-server.db
```