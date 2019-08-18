# Todo
* Dump all lists to "raw-data/data/"
* Create content files based on uuid in the filename from "raw-data/data/"
* Create the reader of the files to serve the content.

# Run server

```
go run cmd/server/main.go -static="/Users/tinkerbell/git/learnalist-api/alists/site-cache"
```

# Get all my lists
```
cd server/alists/
mkdir -p ./site-cache/
mkdir -p ./hugo/content/alists/
mkdir -p ./hugo/data/lists/
rm ./hugo/content/alists/*
rm ./hugo/data/lists/*
go run cmd/get-all-my-lists/main.go -basic-auth="XXX=" -data-dir="$(pwd)/hugo/data/lists/"
go run cmd/make-content/main.go --data-dir="$(pwd)/hugo/data/lists/" --content-dir="$(pwd)/hugo/content/alists/"
go run cmd/build-static/main.go -static="$(pwd)/hugo"
```

```
cd server/alists/
mkdir -p ./site-cache/
mkdir -p ./hugo/content/alists/
mkdir -p ./hugo/data/lists/
rm ./hugo/content/alists/*
rm ./hugo/data/lists/*
go run cmd/get-all-my-lists/main.go -basic-auth="XXX=" -data-dir="$(pwd)/hugo/data/lists/"
go run cmd/make-content/main.go --data-dir="$(pwd)/hugo/data/lists/" --content-dir="$(pwd)/hugo/content/alists/"
go run cmd/build-static/main.go -static="$(pwd)/hugo"
cp -r ./hugo/public-alist/* ./site-cache/
go run cmd/server/main.go -static="/Users/tinkerbell/git/learnalist-api/alists/site-cache"
```

# Process all the items, acting like a full rebuild.
```
rm ./content/alists/*
rm ./data/lists/*
cp ./raw-data/data/* ./data/lists/
go run make-content.go
hugo --cleanDestinationDir -e alist --config=config/alist/config.toml
cp -r ./public-alist/* ./site-cache/
```


# Process a single list, acting like an update of one list on the fly.
```
rm ./content/alists/*
rm ./data/lists/*
cp ./raw-data/data/959170a5-c5c1-5272-ac85-232ff81f9c16.json ./data/lists/
go run make-content.go -uuid="959170a5-c5c1-5272-ac85-232ff81f9c16"
hugo --cleanDestinationDir -e alist --config=config/alist/config.toml
cp -r ./public-alist/* ./site-cache/
```



# Get a single list
```
curl -XGET 'https://learnalist.net/api/v1/alist/3cf1f1ef-44f6-5298-b829-fdaa39919e4a' -u'USER:PASSWORD' > raw-data/data/3cf1f1ef-44f6-5298-b829-fdaa39919e4a.json
```


# Get my lists
```
curl -XGET 'https://learnalist.net/api/v1/alist/by/me' -u'USER:PASSWORD' > me.json
```

# Reference
- [curl-to-go](https://mholt.github.io/curl-to-go)
- [learnalist api docs](https://github.com/freshteapot/learnalist-api/tree/master/api/doc)
- [learnalist docs on api](https://github.com/freshteapot/learnalist-api/blob/master/api/doc/api.md)
