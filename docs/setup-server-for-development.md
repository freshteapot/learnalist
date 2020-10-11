# Install server for development

## Setup
- setup folder structure and remove any data that already exists
- setup empty database

```sh
make clear-site
make rebuild-db
```

##  Run the server
-
```sh
make run-api-server
```

## Run the server and run hugo from within
```sh
HUGO_EXTERNAL=false make run-api-server
```

## Run hugo, server, js
- this will use hugo externally
- hugo on port 1313
- server on port 1234
```sh
make develop
```


## Run via docker
```
make clear-site
make rebuild-db
make build-image-base
make build-image
```

```
docker run --rm --name learnalist \
-v $(pwd)/hugo:/srv/learnalist/hugo \
-v $(pwd)/config:/srv/learnalist/config \
-v /tmp/learnalist/server.db:/srv/learnalist/server.db \
-p 1234:1234 \
-e HUGO_EXTERNAL=false \
learnalist:latest --config=/srv/learnalist/config/docker.config.yaml server
```

# Develop with nats
## Run nats + stan
```sh
docker run \
-p 4222:4222 \
-p 8222:8222 \
-v /tmp/nats-store/:/tmp/nats-store/ nats-streaming:alpine3.12 \
--max_age 10s \
--store=FILE \
--dir=/tmp/nats-store \
--file_auto_sync=1ms \
--stan_debug=true \
--debug=true \
--http_port 8222
```

## Run development
```sh
make clear-site rebuild-db
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTERID="test-cluster" \
EVENTS_STAN_CLIENTID="lal-server" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make develop
```
