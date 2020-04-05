
run-api-server:
	cd server && \
	go run --tags="json1" main.go --config=../config/dev.config.yaml server
