# Install server for development

## Setup folder structure

```sh
rm -rf /tmp/learnalist/
mkdir -p /tmp/learnalist/site-cache
```

## Empty the public directory

```sh
rm -rf ./hugo/public
```

## Make required folders and copy static files to the site-cache

```sh
mkdir -p ./hugo/{public,content/alist,content/alistsbyuser,data/alist,data/alistsbyuser}
cp -rf ./hugo/static/ /tmp/learnalist/site-cache/
```

##  Build the database
```sh
make rebuild-db
```


##  Run the server
```sh
make run-api-server
```


## Run the js + hugo + server
```sh
cd js/
npm install
npm run dev
```

