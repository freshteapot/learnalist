# Learnalist - Education by one list at a time.

# Today
[Vaporware](https://en.wikipedia.org/wiki/Vaporware).
Check [status.json](./status.json) for current status. (Not very useful, yet...)

# Tomorrow

A way to learn via "alist". Made by you, another human or something else.
It will be a service, which will consume the Learnalist API. Hosted via learnalist.net or privately.


# Getting Started

Grab the repo.
```
git clone https://github.com/freshteapot/learnalist.git
cd learnalist/api
go get .
go run api/main.go --port=1234 --database=/tmp/api.db
```
Your server should now be running on port 1234 with the database created at /tmp/api.db

```
curl -i http://localhost:1234
```

Should produce something like
```
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Sat, 24 Sep 2016 14:24:46 GMT
Content-Length: 31

{"message":"1, 2, 3. Lets go!"}
```

When the database is created, two types of lists are added.
You can query all (not quite right yet as it should be linked to an uuid):
```
curl http://localhost:1234/alist/by/me
```
Or an individual list.
```
curl http://localhost:1234/alist/efeb4a6e-9a03-5aff-b46d-7f2ba1d7e7f9
```

## List types

### V1

```
curl -XPOST http://localhost:1234/alist/ -d'
{
        "data": [
            "a",
            "b"
        ],
        "info": {
            "title": "I am a list",
            "type": "v1"
        },
        "uuid": "230bf9f8-592b-55c1-8f72-9ea32fbdcdc4"
    }
'
```
{
    "data": {
        "car": "bil",
        "water": "vann"
    },
    "info": {
        "title": "I am a list with items",
        "type": "v2"
    },
    "uuid": "efeb4a6e-9a03-5aff-b46d-7f2ba1d7e7f9"
}


# Api

| Method | Uri | Description |
| --- | --- | --- |
| POST | /alist | Save a list. |
| PATCH | /alist/{uuid} | Update one or more fields to the list. |
| PUT | /alist/{uuid} | Update all fields allowed to a list. |
| GET | /alist/{uuid} | Get a list via uuid. |
| GET | /alist/by/{uuid} | Get lists by {uuid}. Allow for both public, private lists. |



# References as this becomes more useful.

* https://echo.labstack.com/
* https://github.com/thewhitetulip/web-dev-golang-anti-textbook
* https://gobyexample.com/command-line-flags
* https://developer.github.com/v3/
* [Example that helped understand Unmarshall and Marshall 1](http://mattyjwilliams.blogspot.no/2013/01/using-go-to-unmarshal-json-lists-with.html)
* [Example that helped understand Unmarshall and Marshall 2](https://gist.github.com/mdwhatcott/8dd2eef0042f7f1c0cd8)

# References as I dive deeper into golang.
* https://gobyexample.com/json
* [Like casting but not](https://golang.org/ref/spec#Type_assertions)
* Interfaces http://go-book.appspot.com/interfaces.html
