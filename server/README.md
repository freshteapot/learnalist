# Learnalist - API

## How ugly is the code?
![Code coverage, manually ran](./coverage_badge.png) <a href="https://goreportcard.com/report/github.com/freshteapot/learnalist-api" target="_blank">learnalist-api on goreportcard.</a> (In a new window).

# Some documentation
* [Api](./doc/api.md)
* [List types overview](./doc/list.types.md)
* [Question and Answers](./doc/qa.md)
* [manual install instructions for me](./doc/INSTALL.md)
* [client commands](./doc/client.md)
* [golang tips](./doc/tips.md)
* [Try curl requests](./doc/play.along.md)

# Getting Started

* Grab the repo
```sh
git clone https://github.com/freshteapot/learnalist-api.git
cd learnalist-api/server
```
Then we need to fake a few things whilst go-plus improves its handling of go modules
```sh
GO111MODULE=on go mod init
GO111MODULE=on go mod vendor
```

Build the database
```
ls db/*.sql | sort | xargs cat | sqlite3 /tmp/server.db
```

Now we can run the app
```sh
go run commands/api/main.go \
--port=1234 \
--database=/tmp/server.db \
--hugo-dir="/Users/tinkerbell/git/learnalist-api/server/alists/hugo" \
--site-cache-dir="/Users/tinkerbell/git/learnalist-api/alists/site-cache"

```
Your server should now be running on port 1234 with the database created at /tmp/api.db


# Build for server
Create an apiserver binary including variables injected in during the build step.
```sh
sh build.sh
```

# Once the binary is running.
```sh
curl -i http://localhost:1234/api/v1/
```

Should produce something like
```sh
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 24 Sep 2016 14:24:46 GMT
Content-Length: 31

{"message":"1, 2, 3. Lets go!"}
```

Go [try some curl requests.](./doc/play.along.md)


# References as this becomes more useful.

* https://echo.labstack.com/
* Managing dependencies with [govendor](https://github.com/kardianos/govendor)
* https://github.com/thewhitetulip/web-dev-golang-anti-textbook
* https://gobyexample.com/command-line-flags
* https://developer.github.com/v3/
* [Example that helped understand Unmarshall and Marshall 1](http://mattyjwilliams.blogspot.no/2013/01/using-go-to-unmarshal-json-lists-with.html)
* [Example that helped understand Unmarshall and Marshall 2](https://gist.github.com/mdwhatcott/8dd2eef0042f7f1c0cd8)
* [Getting a date in a format I can parse to the bash script](https://stackoverflow.com/questions/21363187/git-show-dates-in-utc)
* [How to add build time variables to the go application](https://github.com/Ropes/go-linker-vars-example)
* [Parse curl commands and get GO back](https://mholt.github.io/curl-to-go)

# References as I dive deeper into golang.
* https://gobyexample.com/json
* [Like casting but not](https://golang.org/ref/spec#Type_assertions)
* Interfaces http://go-book.appspot.com/interfaces.html
* Pretty gopher badge at the top(https://github.com/jpoles1/gopherbadger).

# Problems

* Slow to run via 'go run'
```sh
cd ./vendor/github.com/mattn/go-sqlite3/
go install
```

Thanks to http://stackoverflow.com/a/38296407.
