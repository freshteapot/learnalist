# Setting up learnalist on kubernetes

# Note
- Make sure rsync is installed on both local and remote.
- followed [setup kubernetes with k3s](./k3s-setup.md).

```sh
export KUBECONFIG="/Users/tinkerbell/.k3s/lal01.learnalist.net.yaml"
export SSH_SERVER="lal01.learnalist.net"
```

# Setting up https
- I used [acme.sh](https://github.com/acmesh-official/acme.sh), to set up the domain via my dns provider.

```sh
acme.sh --issue -d learnalist.net  -d '*.learnalist.net'  --dns dns_namecom
```
I did this as cert-manager just seemed a lot of overhead, once I add a cronjob to run regularly, this problem will be solved.

## Using cronjob
- didnt work... but the commands

```sh
cd ~/git/learn-acme.sh
helm repo add certs https://math-nao.github.io/certs/charts
helm fetch --untar certs/certs
mkdir -p output
helm template certs   --name certs   --values values.yaml   --output-dir=./output
```

## How to create the secret used for tls
- Didnt work via https://github.com/math-nao/certs, but maybe this was due to multiple domains.
  It worked, it just didnt apply it to the server correcly.
- Running it manually and replacing the file worked great

```sh
kubectl create secret tls tls \
--key /Users/tinkerbell/.acme.sh/learnalist.net/learnalist.net.key \
--cert /Users/tinkerbell/.acme.sh/learnalist.net/fullchain.cer
```


```sh
kubectl create secret tls tls \
--key /Users/tinkerbell/.acme.sh/learnalist.net/learnalist.net.key \
--cert /Users/tinkerbell/.acme.sh/learnalist.net/fullchain.cer --dry-run -o yaml | kubectl apply -f -
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
// TODO is this needed now?
// Check via k3d or something
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
