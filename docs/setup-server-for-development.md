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
## Run nats locally

```sh
make run-nats-from-docker
```

## Running the api server

```sh
make clear-site rebuild-db
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-server" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make run-api-server
```

## Running the development locally
### nats
```sh
make clear-site rebuild-db
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-server" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make develop
```
### memory
```sh
make clear-site rebuild-db
EVENTS_VIA="memory" \
HUGO_EXTERNAL=false \
make develop-localhost
```

## Running slack events
- Get slack secret from the cluster, checkout [api.events](./api.events.md)

```sh
cd server
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-slack-events" \
EVENTS_NATS_SERVER="127.0.0.1" \
EVENTS_SLACK_WEBHOOK="XXX" \
go run main.go --config=../config/dev.config.yaml tools slack-events
```


## Running the event reader

```sh
cd server
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-event-reader" \
EVENTS_NATS_SERVER="127.0.0.1" \
go run main.go --config=../config/dev.config.yaml tools event-reader
```


## Run the challenge sync service
```sh
TOPIC=lal.monolog \
EVENTS_STAN_CLIENT_ID=challenges-sync \
EVENTS_STAN_CLUSTER_ID=test-cluster \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools challenge sync
```

## Read topic
### lal.monolog
Main topic where almost all events go
### notifications
Topic where communications goto

```sh
TOPIC=lal.monolog \
EVENTS_STAN_CLIENT_ID=nats-reader \
EVENTS_STAN_CLUSTER_ID=test-cluster \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools natsutils read
```

