GO111MODULE=on go test -v `go list ./... | grep -vE 'event|integrations'` -covermode=count -coverprofile=profile.cov
GO111MODULE=on go tool cover -html=profile.cov
