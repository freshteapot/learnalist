# Learnalist - API

* [manual install instructions for me](./doc/INSTALL.md)
* [client commands](./doc/client.md)

# Getting Started

* Grab the repo
```sh
git clone https://github.com/freshteapot/learnalist.git
cd learnalist/api
```
Then we need to fake a few things whilst go-plus improves its handling of go modules
```sh
GO111MODULE=on go mod init
GO111MODULE=on go mod vendor
```
Now we can run the app
```sh
go run commands/api/main.go --port=1234 --database=/tmp/api.db
```
Your server should now be running on port 1234 with the database created at /tmp/api.db


# Build for server
Create an apiserver binary including variables injected in during the build step.
```sh
sh build.sh
```

# Once the binary is running.
```sh
curl -i http://localhost:1234
```

Should produce something like
```sh
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 24 Sep 2016 14:24:46 GMT
Content-Length: 31

{"message":"1, 2, 3. Lets go!"}
```

## Play along.
When the database is created, it is empty.

## You need a user first.
```sh
curl -XPOST 127.0.0.1:1234/register -d'
{
    "username":"chris",
    "password":"test"
}
'
```

### Add a list of type v1.

```sh
curl -XPOST http://localhost:1234/alist -u'chris:test' -d'
{
    "data": [
        "monday",
        "tuesday",
        "wednesday",
        "thursday",
        "friday",
        "saturday",
        "sunday"
    ],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}
'
```

### Get all lists created by you.
```sh
curl http://localhost:1234/alist/by/me
```

### Add a list of type v2.

```sh
curl -XPOST http://localhost:1234/alist -uchris:test -d'
{
    "data": [
        {
            "from": "chris",
            "to": "christopher"
        }
    ],
    "info": {
        "title": "A list of key:value pairs.",
        "type": "v2"
    }
}'
```

Again, query all the lists by you.
```sh
curl http://localhost:1234/alist/by/me
```

Or an individual list.
```sh
curl http://localhost:1234/alist/{uuid}
```

# Api

| Method | Uri | Description |
| --- | --- | --- |
| POST | /alist | Save a list. |
| PATCH | /alist/{uuid} | Update one or more fields to the list. |
| PUT | /alist/{uuid} | Update all fields allowed to a list. |
| GET | /alist/{uuid} | Get a list via uuid. |
| GET | /alist/by/{uuid} | Get lists by {uuid}. Allow for both public, private lists. |



# List types

| Type | Description |
| --- | --- |
| v1 | An array of a string.|
| v2 | An array of object alist.AlistItemTypeV2 |

### V1

```
{
    "data": [
        "a",
        "b"
    ],
    "info": {
        "title": "A list of strings",
        "type": "v1"
    }
}
'
```

### V2

```
{
    "data": [
        {
            "from": "chris",
            "to": "chris"
        }
    ],
    "info": {
        "title": "A list of key:value pairs.",
        "type": "v2"
    }
}
```

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

# References as I dive deeper into golang.
* https://gobyexample.com/json
* [Like casting but not](https://golang.org/ref/spec#Type_assertions)
* Interfaces http://go-book.appspot.com/interfaces.html


# Problems

* Slow to run via 'go run'
```sh
cd ./vendor/github.com/mattn/go-sqlite3/
go install
```

Thanks to http://stackoverflow.com/a/38296407.

* Update all vendors
```sh
govendor fetch +v
```


# Working with structs and json

Get the Data object and add a single row to the v2 type data.
```go
aListV2Data := aList.Data.(alist.AlistTypeV2)

item := &alist.AlistItemTypeV2{From: "Hi", To: "Hello"}
aListV2Data = append(aListV2Data, *item)
aList.Data = aListV2Data
```
