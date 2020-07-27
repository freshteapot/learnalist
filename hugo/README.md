# Add test files
```sh
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.json data/alist/
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.md content/alist/
hugo server -e alist --config=config/alist/config.toml -w
```

# View hugo server
```sh
http://localhost:1313/alist/5d4c9869-1d26-567d-82be-497c3521368a.html
```

# Make list public
```sh
curl -XPUT 'http://localhost:1234/api/v1/share/readaccess' -u'iamchris:test123' -d '{
  "alist_uuid": "5d4c9869-1d26-567d-82be-497c3521368a",
  "action": "public"
}'
```

# View via the server
```sh
http://localhost:1234/alist/5d4c9869-1d26-567d-82be-497c3521368a.html
```


docker run --name lal-sample \
-p 8080:80 \
-v $PWD/hugo/public:/usr/share/nginx/html:ro \
-P -d nginx:1.17-alpine
