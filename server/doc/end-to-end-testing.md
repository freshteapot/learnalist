# End to end testing

# Setup the Server
```sh
rm -rf /tmp/learnalist-api
mkdir -p /tmp/learnalist-api/site-cache
```

```sh
cd server/
```

# Build the database
```sh
ls db/*.sql | sort | xargs cat | sqlite3 /tmp/learnalist-api/server.db
```


# Start the server
```sh
go run commands/api/main.go \
--port=1234 \
--database=/tmp/learnalist-api/server.db \
--hugo-dir="$(pwd)/../hugo" \
--site-cache-dir="/tmp/learnalist-api/site-cache"
```

# Run all tests
Adding the clean testcache, makes sure it reconnects via http.

```
cd e2e
go clean -testcache && go test -test.v .
```

```
go clean -testcache && go test -test.v -run="TestUserHasTwoLists" .
```
