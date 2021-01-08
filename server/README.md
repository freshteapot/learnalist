# Learnalist - API

## Requirements
- hugo to be installed (> v0.74.3)
- golang (> go1.15.2)
- nodejs

## How ugly is the code?

```sh
cd server
goreportcard-cli -v
```
Via https://github.com/gojp/goreportcard#command-line-interface

OR

```sh
cd server
staticcheck ./...
```

# Some documentation
* [Api](../doc/api.auto.md)
* [List types overview](../doc/list.types.md)
* [Question and Answers](../doc/qa.md)
* [manual install instructions for me](../doc/INSTALL.md)
* [client commands](../doc/client.md)
* [golang tips](../doc/tips.md)
* [Try curl requests](../doc/play.along.md)
* [Explaining the sharing and access control](../doc/sharing.md)

# Getting Started
[Setting up for development](../docs/setup-server-for-development.md).

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

