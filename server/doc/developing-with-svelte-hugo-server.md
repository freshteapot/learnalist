# Server

Setup folder structure
```sh
rm -rf /tmp/learnalist-api
mkdir -p /tmp/learnalist-api/site-cache
```

```sh
rm -rf ./hugo/public-alist
```

```sh
mkdir -p ./hugo/{public-alist,content/alists,data/lists}
cp -rf ./hugo/themes/alist/static/ /tmp/learnalist-api/site-cache/
```

# Build the database
```sh
ls server/db/*.sql | sort | xargs cat | sqlite3 /tmp/learnalist-api/server.db
```


# Run the server
```sh
cd server/
```


```sh
go run commands/api/main.go \
--port=1234 \
--database=/tmp/learnalist-api/server.db \
--hugo-dir="$(pwd)/../hugo" \
--site-cache-dir="/tmp/learnalist-api/site-cache"
```

# Rebuild from existing database

```
go run commands/rebuild-static-site/main.go \
--database=/tmp/learnalist-api/server.db \
--hugo-dir="$(pwd)/../hugo" \
--site-cache-dir="/tmp/learnalist-api/site-cache"
```

# Svelte
```
cd svelte
```

# Copy to themes
```sh
npm run build
cp public/v1.js ../hugo/themes/alist/static/js/
cp public/v1.js.map ../hugo/themes/alist/static/js/
cp public/user.js ../hugo/themes/alist/static/js/
cp public/user.js.map ../hugo/themes/alist/static/js/
cp public/css/tachyons.min.css ../hugo/themes/alist/static/css/tachyons.min.css

cd ../hugo/
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.json data/lists/
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.md content/alists/
cd -
```

# Run hugo only
```sh
cd hugo
hugo server -e alist --config=config/alist/config.toml -w
```

# Use a test to make a list and set it to public
```sh
go test -run TestSharePublic2 -v .
```
