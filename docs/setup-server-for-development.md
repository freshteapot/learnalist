# Install server for development

```sh
git clone https://github.com/freshteapot/learnalist-api.git
cd learnalist-api/
```

# Zero to hero
## Start nats
```sh
make run-nats-from-docker
```

## Run
### Full stack
- useful when you want to work with JS or the UI

```sh
make clear-site rebuild-db
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-server" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make develop
```

Your server should now be running on port 1234 with the database created at /tmp/learnalist/server.db.

```sh
curl -i http://localhost:1234/api/v1/
```

Go [try some curl requests.](./play.along.md)



### Api
- useful when you want to work with JS or the UI

```sh
make clear-site rebuild-db
EVENTS_VIA="nats" \
EVENTS_STAN_CLUSTER_ID="test-cluster" \
EVENTS_STAN_CLIENT_ID="lal-server" \
EVENTS_NATS_SERVER="127.0.0.1" \
HUGO_EXTERNAL=false \
make run-api-server
```

# More details

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

## Run via docker
```sh
make clear-site
make rebuild-db
make build-image-base
make build-image
```

```sh
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
## Running slack events
- Via [slack website](https://api.slack.com)
- Get slack secret from the cluster ([api.events](./api.events.md))

```sh
EVENTS_SLACK_WEBHOOK="XXX" \
scripts/run-slack.sh
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
EVENTS_VIA="nats" \
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
EVENTS_VIA="nats" \
EVENTS_STAN_CLIENT_ID=nats-reader \
EVENTS_STAN_CLUSTER_ID=test-cluster \
EVENTS_NATS_SERVER=127.0.0.1 \
go run main.go --config=../config/dev.config.yaml \
tools natsutils read
```
