# Setting up learnalist on kubernetes

# Note
- Make sure rsync is installed on both local and remote


# Setting up https
- I used [acme.sh](https://github.com/acmesh-official/acme.sh), to set up the domain via my dns provider.

```sh
acme.sh --issue -d learnalist.net  -d '*.learnalist.net'  --dns dns_namecom
```
I did this as cert-manager just seemed a lot of overhead, once I add a cronjob to run regularly, this problem will be solved.

## Copy the keys to the server

```sh
scp /Users/tinkerbell/.acme.sh/learnalist.net/* learnalist.net:/
```

## How to create the secret used for tls

```sh
kubectl create secret tls tls \
--key /Users/tinkerbell/.acme.sh/learnalist.net/learnalist.net.key \
--cert /Users/tinkerbell/.acme.sh/learnalist.net/learnalist.net.cer
```

```
export SSH_SERVER="lal01.learnalist.net"
```

# Setup insecure local registry
```sh
rsync -avzP --rsync-path="sudo rsync" ./k3s/registries.yaml $SSH_SERVER:/etc/rancher/k3s/registries.yaml
```

```sh
sudo systemctl restart k3s
```

# Setup Configmap
-
## Create
```sh
kubectl create configmap learnalist-config --from-file=config.yaml=config/docker.config.yaml
```

## Update
```sh
kubectl create configmap learnalist-config --from-file=config.yaml=config/docker.config.yaml -o yaml --dry-run | kubectl replace -f -
```

## View
```
kubectl get configmap learnalist-config --from-file=config.yaml=config/docker.config.yaml -o yaml
```



# Sync files
## hugo + javascript
```sh
rsync -avzP \
--rsync-path="sudo rsync" \
--exclude-from="exclude-srv-learnalist.txt" \
./hugo $SSH_SERVER:/srv/learnalist
```

## hugo public files

```sh
rsync -avvvzP \
--rsync-path="sudo rsync" \
--exclude-from='exclude-srv-learnalist-public.txt' \
./hugo/public/ $SSH_SERVER:/srv/learnalist/site-cache
```


# Rebuild static-site
- rebuild all lists
- rebuild all user lists

```sh
/app/bin/learnalist-cli --config=/etc/learnalist/config.yaml tools rebuild-static-site
```


# Reference
- [acme.sh](https://github.com/acmesh-official/acme.sh)
