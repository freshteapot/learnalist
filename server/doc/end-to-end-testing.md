# End to end testing
## Get the server up and running

[Setup the server for development](./install-server-for-dev.md)


# Run all tests
Adding the clean testcache, makes sure it reconnects via http.

```
cd e2e
go clean -testcache && go test -test.v .
```

```
go clean -testcache && go test -test.v -run="TestUserHasTwoLists" .
```
