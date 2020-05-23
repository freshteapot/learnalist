# End to end testing
## Get the server up and running

[Setup the server for development](./install-server-for-dev.md)


# Run all tests
Adding the clean testcache, makes sure it reconnects via http.

- need to skip, due to the race conditions
```
cd server/e2e
go clean -testcache && go test  -ginkgo.v -ginkgo.progress -ginkgo.skip="Smoke list access"  -test.v .
go clean -testcache && go test  -ginkgo.v -ginkgo.progress -ginkgo.focus="Smoke list access"  -test.v .
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
go clean -testcache && go test  -ginkgo.v -ginkgo.progress -ginkgo.focus="Smoke list access"  -test.v .
```


# Reference
- https://onsi.github.io/ginkgo/#focused-specs
