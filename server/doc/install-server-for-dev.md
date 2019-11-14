# Install server for development

## Setup folder structure

```sh
rm -rf /tmp/learnalist-api
mkdir -p /tmp/learnalist-api/site-cache
```

## Empty the public directory

```sh
rm -rf ./hugo/public-alist
```

## Make required folders and copy static files to the site-cache

```sh
mkdir -p ./hugo/{public-alist,content/alists,data/lists}
cp -rf ./hugo/themes/alist/static/ /tmp/learnalist-api/site-cache/
```

##  Build the database
```sh
ls server/db/*.sql | sort | xargs cat | sqlite3 /tmp/learnalist-api/server.db
```


##  Run the server
```sh
cd server/
```

```sh
go run main.go server --config=../config/dev.config.yaml
```
