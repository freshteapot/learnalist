GIT_COMMIT:=$(shell git rev-parse HEAD)
GIT_HASH_DATE:=$(shell TZ=UTC git show --quiet --date='format-local:%Y%m%dT%H%M%SZ' --format="%cd" ${GIT_COMMIT})

###############################################################################
#
# Development commands
#
clear-site:
	rm -rf ./hugo/public/*
	mkdir -p ./hugo/{public/alist,public/alistsbyuser}
	mkdir -p ./hugo/{content/alist,data/alist}
	mkdir -p ./hugo/{content/alistsbyuser,data/alistsbyuser}
	rm -f ./hugo/content/alist/*
	rm -f ./hugo/content/alistsbyuser/*
	rm -f ./hugo/data/alist/*
	rm -f ./hugo/data/alistsbyuser/*
	echo "[]" > ./hugo/data/public_lists.json
	cd ./hugo && hugo

rebuild-db:
	mkdir -p /tmp/learnalist/
	rm -f /tmp/learnalist/server.db
	ls server/db/*.sql | sort | xargs cat | sqlite3 /tmp/learnalist/server.db

test:
	cd server && \
	./cover.sh

run-api-server:
	cd server && \
	go run --tags="json1" main.go --config=../config/dev.config.yaml server

develop:
	cd js && \
	npm run dev

build-mocks:
	cd server && mockery -all -recursive

run-e2e-tests:
	cd server && \
	./run-e2e.sh

generate-openapi-go:
	rm -rf ./server/pkg/openapi && \
	mkdir -p ./server/pkg/openapi && \
	GO_POST_PROCESS_FILE="/usr/local/bin/gofmt -w" \
	openapi-generator generate -i ./learnalist.yaml -g go -o ./server/pkg/openapi && \
	rm ./server/pkg/openapi/go.mod && \
	rm ./server/pkg/openapi/go.sum

generate-openapi-js:
	rm -rf ./js/src/openapi && \
	mkdir -p ./js/src/openapi && \
	openapi-generator generate -i ./learnalist.yaml -g typescript-fetch -o ./js/src/openapi \
	--additional-properties typescriptThreePlus=true \
	--additional-properties modelPropertyNaming=original \
	--additional-properties enumPropertyNaming=original

generate-openapi-dart:
	rm -rf /tmp/openapi/dart && \
	mkdir -p /tmp/openapi/dart && \
	openapi-generator generate -i ./learnalist.yaml -g dart -o /tmp/openapi/dart \


generate-docs-api-overview:
	cd server && \
	yq r ../learnalist.yaml -j | jq -r -c | go run main.go tools --config=../config/dev.config.yaml docs api-overview > ../docs/api.auto.md

###############################################################################
#
# More production than development
#
rebuild-static-site:
	cd server && \
	go run --tags="json1" main.go tools rebuild-static-site --config=../config/dev.config.yaml

build-site-assets:
	./build.sh

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

build-image:
	cd server && \
	docker build \
	--build-arg GIT_COMMIT="${GIT_COMMIT}" \
	--build-arg GIT_HASH_DATE="${GIT_HASH_DATE}" \
	-t learnalist:latest .

push-image:
	cd server && \
	docker tag learnalist:latest registry.devbox:5000/learnalist:latest
	docker push registry.devbox:5000/learnalist:latest
