#go test -v `go list ./... | grep -vE 'event|integrations'` -covermode=count -coverprofile=profile.cov
go test `go list ./... | grep -vE 'event|integrations'` -covermode=count -coverprofile=profile.cov
go tool cover -func=profile.cov | tail -1
go tool cover -html=profile.cov
gopherbadger -covercmd "go tool cover -func=profile.cov"
