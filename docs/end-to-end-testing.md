# End to end testing

# Bring online
## Get the server up and running

```sh
STATIC_SITE_EXTERNAL=false \
make clear-site rebuild-db develop
```

## Start reminder

```sh
make rebuild-db-remind-manager run-remind-manager
```

# Run all tests
```sh
make run-e2e-tests
```

## Get stats
```sh
cd e2elog/example
```
### Get all openapi endpoints that were used
```sh
go run main.go -logs=/tmp/learnalist/e2e-all.log | jq -c '.endpoints[]|select(.touched)'
```

### Get stats
```sh
go run main.go -logs=/tmp/learnalist/e2e-all.log -stats
```

# Manually run tests
Adding the clean testcache, makes sure it reconnects via http.

- need to skip, due to the race conditions
```
cd server/e2e
go clean -testcache && go test --tags="json1" -ginkgo.v -ginkgo.progress -ginkgo.skip="Smoke list access"  -test.v .
go clean -testcache && go test --tags="json1" -ginkgo.v -ginkgo.progress -ginkgo.focus="Smoke list access"  -test.v .
```

```
go clean -testcache && go test -test.v -run="TestUserHasTwoLists" .
```

# Use a test to make a list and set it to public
```sh
go test -run TestSharePublic2 -v .
```

### Run specific test
```
go clean -testcache && go test --tags="json1" -ginkgo.v -ginkgo.progress -ginkgo.focus="Smoke list access"  -test.v .
```

```
go clean -testcache && go test --tags="json1" -ginkgo.v -ginkgo.progress -ginkgo.focus="Smoke user"  -test.v .
```

Smoke tests, which dont clean up
```
make run-e2e-tests
cd -
sqlite3 /tmp/learnalist/server.db  'SELECT uuid FROM user' | STATIC_SITE_EXTERNAL=false  xargs -I {}  go run --tags="json1" main.go tools user delete --config=../config/dev.config.yaml --dsn=/tmp/learnalist/server.db {}
```


# Reference
- https://onsi.github.io/ginkgo/#focused-specs
