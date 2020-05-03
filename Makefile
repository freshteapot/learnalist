
clear-site:
	mkdir -p ./hugo/{public,content/alist,data/alist,content/alistsbyuser,data/alistsbyuser}
	rm -rf ./hugo/public/*
	rm -f ./hugo/content/alist/*
	rm -f ./hugo/content/alistsbyuser/*
	rm -f ./hugo/data/alist/*
	rm -f ./hugo/data/alistsbyuser/*
	echo "[]" > ./hugo/data/public_lists.json

rebuild-db:
	mkdir -p /tmp/learnalist/
	rm -f /tmp/learnalist/server.db
	ls server/db/*.sql | sort | xargs cat | sqlite3 /tmp/learnalist/server.db

tests:
	cd server && \
	./cover.sh

run-api-server:
	cd server && \
	go run --tags="json1" main.go --config=../config/dev.config.yaml server

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

develop:
	cd js && \
	npm run dev
