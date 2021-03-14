GIT_COMMIT:=$(shell git rev-parse HEAD)
GIT_HASH_DATE:=$(shell TZ=UTC git show --quiet --date='format-local:%Y-%m-%dT%H:%M:%SZ' --format="%cd" ${GIT_COMMIT})
###############################################################################
#
# Development commands
#
clear-site:
	rm -rf ./hugo/public/*
	mkdir -p ./hugo/{public/alist,public/alistsbyuser}
	mkdir -p ./hugo/{content/alist,data/alist}
	mkdir -p ./hugo/{content/alistsbyuser,data/alistsbyuser}
	mkdir -p ./hugo/{content/challenge,data/challenge}
	rm -f ./hugo/content/alist/*
	rm -f ./hugo/data/alist/*
	rm -f ./hugo/content/alistsbyuser/*
	rm -f ./hugo/data/alistsbyuser/*
	rm -f ./hugo/content/challenge/*
	rm -f ./hugo/data/challenge/*
	echo "[]" > ./hugo/data/public_lists.json
	cd ./hugo && hugo

rebuild-db:
	mkdir -p /tmp/learnalist/
	rm -f /tmp/learnalist/server.db
	ls server/db/*.sql | sort | xargs cat | sqlite3 /tmp/learnalist/server.db

rebuild-db-remind-manager:
	mkdir -p /tmp/learnalist/
	rm -f /tmp/learnalist/remind-daily.db
	ls server/db/*.sql | sort | xargs cat | sqlite3 /tmp/learnalist/remind-daily.db

rebuild-static-site:
	cd server && \
	EVENTS_STAN_CLIENT_ID=rebuild-static-site \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run --tags="json1" main.go tools rebuild-static-site --config=../config/dev.config.yaml

test:
	cd server && \
	./cover.sh

run-nats-from-docker:
	docker rm -f lal-nats &>/dev/null && \
	rm -rf /tmp/nats-store/ && \
	mkdir -p /tmp/nats-store/ && \
	docker run \
	-d \
	--rm \
	--name lal-nats \
	-p 4222:4222 \
	-p 8222:8222 \
	-v /tmp/nats-store/:/tmp/nats-store/ nats-streaming:alpine3.12 \
	--store=FILE \
	--dir=/tmp/nats-store \
	--stan_debug=true \
	--debug=true \
	--http_port 8222

run-challenges-sync:
	cd server && \
	TOPIC=lal.monolog \
	EVENTS_STAN_CLIENT_ID=challenges-sync \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run main.go --config=../config/dev.config.yaml \
	tools challenges sync

run-notifications-push-notifications:
	cd server && \
	TOPIC=notifications \
	EVENTS_STAN_CLIENT_ID=notifications-push-notifications \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run main.go --config=../config/dev.config.yaml \
	tools notifications push-notifications

run-api-server:
	cd server && \
	EVENTS_STAN_CLIENT_ID=lal-server \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run --tags="json1" main.go --config=../config/dev.config.yaml server

run-remind-manager:
	cd server && \
	TOPIC=lal.monolog \
	EVENTS_STAN_CLIENT_ID=remind-daily \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run --tags=json1 main.go --config=../config/dev.config.yaml \
	tools remind manager

run-static-site:
	cd server && \
	EVENTS_STAN_CLIENT_ID=static-site \
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	go run main.go --config=../config/dev.config.yaml \
	static-site

# Running development with hugo and golang ran outside of the javascript landscape
# Enables the ability to expose the code to my ip address not just localhost
develop:
	EVENTS_STAN_CLUSTER_ID=test-cluster \
	EVENTS_NATS_SERVER=127.0.0.1 \
	scripts/run-hugo.sh

develop-localhost:
	cd js && \
	npm run dev

build-mocks:
	cd server && mockery --all --recursive

run-e2e-tests:
	cd server && \
	./run-e2e.sh

generate-openapi-one:
	rm -rf /tmp/openapi/one && \
	mkdir -p /tmp/openapi/one && \
	cp ./openapi/learnalist.yaml /tmp/openapi/one/ && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.version.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.asset.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.plank.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.user.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.alist.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.spaced_repetition.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.challenge.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.mobile.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.remind.yaml && \
	yq m -i /tmp/openapi/one/learnalist.yaml ./openapi/api.app_settings.yaml && \
	openapi-generator generate -i /tmp/openapi/one/learnalist.yaml -g openapi-yaml -o /tmp/openapi/one

generate-openapi-markdown: generate-openapi-one
	openapi-generator generate -i /tmp/openapi/one/learnalist.yaml -g markdown -o /tmp/openapi/one

generate-openapi-go: generate-openapi-one
	rm -rf ./server/pkg/openapi && \
	mkdir -p ./server/pkg/openapi && \
	GO_POST_PROCESS_FILE="/usr/local/bin/gofmt -w" \
	openapi-generator generate -i /tmp/openapi/one/learnalist.yaml -g go -o ./server/pkg/openapi && \
	rm ./server/pkg/openapi/go.mod && \
	rm ./server/pkg/openapi/go.sum

generate-openapi-js: generate-openapi-one
	rm -rf ./js/src/openapi && \
	mkdir -p ./js/src/openapi && \
	openapi-generator generate -i /tmp/openapi/one/openapi/openapi.yaml -g typescript-fetch -o ./js/src/openapi \
	--additional-properties typescriptThreePlus=true \
	--additional-properties modelPropertyNaming=original \
	--additional-properties enumPropertyNaming=original

generate-openapi-dart: generate-openapi-one
	rm -rf /tmp/openapi/dart && \
	mkdir -p /tmp/openapi/dart && \
	DART_POST_PROCESS_FILE="/usr/local/bin/dartfmt -w" \
	openapi-generator generate -i /tmp/openapi/one/learnalist.yaml -g dart -o /tmp/openapi/dart \
	--additional-properties ensureUniqueParams=false

generate-docs-api-overview: generate-openapi-one
	cd server && \
	cat /tmp/openapi/one/learnalist.yaml | go run main.go tools --config=../config/dev.config.yaml docs api-overview > ../docs/api.auto.md
###############################################################################
#
# More production than development
#
build-site-assets:
	./scripts/build-site-assets.sh

sync-site-assets:
	rsync -avzP \
	--rsync-path="sudo rsync" \
	--exclude-from="exclude-srv-learnalist.txt" \
	./hugo ${SSH_SERVER}:/srv/learnalist

sync-db-files:
	rsync -avzP \
	--rsync-path="sudo rsync" \
	./server/db ${SSH_SERVER}:/srv/learnalist


build-image-base:
	cd server && \
	docker build -f Dockerfile_prod_base -t learnalist-prod-base:latest .

build-image: generate-openapi-go
	cd server && \
	docker build \
	--build-arg GIT_COMMIT="${GIT_COMMIT}" \
	--build-arg GIT_HASH_DATE="${GIT_HASH_DATE}" \
	-t learnalist:latest .

push-image:
	cd server && \
	docker tag learnalist:latest registry.devbox:5000/learnalist:latest
	docker push registry.devbox:5000/learnalist:latest
