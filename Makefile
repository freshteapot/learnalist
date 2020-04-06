
tests:
	cd server && \
	./cover.sh

run-api-server:
	cd server && \
	go run --tags="json1" main.go --config=../config/dev.config.yaml server

rebuild-static-site:
	cd server && \
	go run --tags="json1" main.go tools rebuild-static-site --config=../config/dev.config.yaml
