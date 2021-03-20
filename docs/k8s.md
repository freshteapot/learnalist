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

## Via K8S
- Works as long as each domain has its own secret in tls

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

### Create
```sh
kubectl create secret tls tls \
--key /Users/tinkerbell/.acme.sh/learnalist.net/learnalist.net.key \
--cert /Users/tinkerbell/.acme.sh/learnalist.net/fullchain.cer
```

### Update
```sh
kubectl create secret tls tls \
--key /Users/tinkerbell/.acme.sh/learnalist.net/learnalist.net.key \
--cert /Users/tinkerbell/.acme.sh/learnalist.net/fullchain.cer --dry-run -o yaml | kubectl apply -f -
```

```sh
export SITE_URL="learnalist.net"
export SITE_SSL_PORT="443"
openssl s_client -connect ${SITE_URL}:${SITE_SSL_PORT} \
  -servername ${SITE_URL} 2> /dev/null |  openssl x509 -noout  -dates
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

```sh
make sync-site-assets
```

# Reference
- [acme.sh](https://github.com/acmesh-official/acme.sh)
- [cronjob to keep ssl updated](https://github.com/math-nao/certs)
- [1 tls per domain](https://github.com/math-nao/certs/issues/36#issuecomment-744317680)
