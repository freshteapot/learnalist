# Sharing svelte store across two customElements
- Built off [svelte-rollup-template from metonym](https://github.com/metonym/svelte-rollup-template).

# What
- shared svelte store
- components use it for dev
- compiled the build step skips adding it to the module

# Show me
- install dependencies
```sh
npm install
```
- build the "superstore".
```sh
rollup -c rollup.config.store.js
```
- run develop
```sh
npm run develop
```
- open localhost:3000
- click on "the count is 0"
- watch both change
- Open developer tools
- type
```sh
superstore.count.subscribe(value => a = value)
```
- type and observe number matches count
```sh
a
```
- click on "the count is X"
- type and observe number matches count
```sh
a
```

# View via docker
## Run
```sh
docker run --name learnalist-static-dev \
-p 8080:80 \
-v $PWD/hugo/public/:/usr/share/nginx/html:ro \
-v $PWD/config/nginx.conf:/etc/nginx/nginx.conf:ro \
-P -d nginx:1.17-alpine
```

## Delete
```sh
docker container rm --force learnalist-static-dev
```


```
cat ~/git/learn-hugo/hugo/list.json | \
jq -c '.[]| .' | go run main.go tools hugo-import-lists \
--config=../config/dev.config.yaml \
--content-dir="/Users/tinkerbell/git/learnalist/hugo/content" \
--data-dir="/Users/tinkerbell/git/learnalist/hugo/data"
```


```
cat ~/git/learn-hugo/hugo/list.json | \
jq -c '.[]| .' \
| go run main.go tools hugo-import-lists-by-user \
--config=../config/dev.config.yaml \
--content-dir="/Users/tinkerbell/git/learnalist/hugo/content" \
--data-dir="/Users/tinkerbell/git/learnalist/hugo/data" \
--user-uuid="fc7f0e39-aa15-52d4-b590-e3a2bf9ee86d"
```


1) Build normal

```sh
hugo server  --environment=dev --config=config/  -v -w --disableFastRender --renderToDisk
```

2) Build and postcss

```sh
cd hugo && HUGO_BUILD_WRITESTATS=true HUGO_PARAMS_BUILDCSS=true HUGO_PARAMS_BUILDCSSPRODUCTION=true hugo --environment=lal01
```

3) build with the manifest css

```sh
cd js && node watch.js
```


```sh
rm -rf ./hugo/public/*
cd hugo && HUGO_PARAMS_BUILDCSS=false hugo --environment=lal01
```


Need test data, to cover all paths thru the templates...

save to static
rsync static
rebuild



```sh
make sync-site-to-k8s
```
