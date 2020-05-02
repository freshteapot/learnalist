# Install onto kubernetes
# Notes
- You will need to add registry.devbox to your /etc/hosts

```sh
export KUBECONFIG="/Users/tinkerbell/.k3s/lal01.learnalist.net.yaml"
export SSH_SERVER="lal01.learnalist.net"
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
/app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools rebuild-static-site
```
