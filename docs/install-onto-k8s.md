# Install onto kubernetes
# Notes
- You will need to add registry.devbox to your /etc/hosts

```
127.0.0.1 registry.devbox
```

# Fire up the registry
- make sure container-registry (docker-registry) is running
- log onto the server and port-forward 5000

```sh
ssh $SSH_SERVER sudo kubectl port-forward deployment/container-registry 5000:5000 &
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
