# Get the server up and running

[Setup the server for development](./install-server-for-dev.md)

# Rebuild from existing database

```sh
go run main.go tools rebuild-static-site --config=dev.config.yaml
```

# Svelte
```sh
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
