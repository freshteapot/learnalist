# Learnalist - API

## How ugly is the code?
![Code coverage, manually ran](./coverage_badge.png) <a href="https://goreportcard.com/report/github.com/freshteapot/learnalist-api" target="_blank">learnalist-api on goreportcard.</a> (In a new window).

# Some documentation
* [Question and Answers](./doc/qa.md)
* [manual install instructions for me](./doc/INSTALL.md)
* [client commands](./doc/client.md)
* [golang tips](./doc/tips.md)
* [Try curl requests](./doc/play.along.md)

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
curl -i http://localhost:1234/v1/
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

# Api

| Method | Uri | Description |
| --- | --- | --- |
| GET | /v1/ | Replies with a simple message. |
| GET | /v1/version | Version informationn about the server. |
| POST | /v1/alist | Save a list. |
| DELETE | /v1/alist/{uuid} | Delete a list via uuid. |
| PUT | /v1/alist/{uuid} | Update all fields allowed to a list. |
| GET | /v1/alist/{uuid} | Get a list via uuid. |
| GET | /v1/alist/by/me(?labels=,list_type={v1,v2}) | Get lists by the currently logged in user. |
| POST | /v1/labels | Save a new label. |
| GET | /v1/labels/by/me | Get labels by the currently logged in user. |
| DELETE | /v1/labels/{uuid} | Delete a label via uuid. |
| POST | /v1/share/alist | Share a list with another user. |



# List types

| Type | Description |
| --- | --- |
| v1 | An array of a string.|
| v2 | An array of object alist.AlistItemTypeV2 |
| v3 | Record your rowing data from a concept2. TypeV3, made up of an array of alist.TypeV3Item |
| v4 | Record content and its url / reference. TypeV4, made up of an array of alist.TypeV4Item|

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

### V3
```
{
  "data": [
    {
      "when": "2019-05-06",
      "overall": {
        "time": "7:15.9",
        "distance": 2000,
        "spm": 28,
        "p500": "1:48.9"
      },
      "splits": [
        {
          "time": "1.46.4",
          "distance": 500,
          "spm": 29,
          "p500": "1:58.0"
        }
      ]
    }
  ],
  "info": {
      "title": "A list to record rows.",
      "type": "v3"
  }
}
```

### V4
```
{
  "data": [
    {
      "content": "Im tough, Im ambitious, and I know exactly what I want. If that makes me a bitch, okay. ― Madonna",
      "url":  "https://www.goodreads.com/quotes/54377-i-m-tough-i-m-ambitious-and-i-know-exactly-what-i"
    },
    {
      "content": "Design is the art of arranging code to work today, and be changeable forever. – Sandi Metz",
      "url":  "https://dave.cheney.net/paste/clear-is-better-than-clever.pdf"
    }
  ],
  "info": {
      "title": "A list of fine quotes.",
      "type": "v4"
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

* Update all vendors
```sh
govendor fetch +v
```
