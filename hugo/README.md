# Add test files
```sh
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.json data/lists/
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.md content/alists/
hugo server -e alist --config=config/alist/config.toml -w
```

# View hugo server
```sh
http://localhost:1313/alists/5d4c9869-1d26-567d-82be-497c3521368a.html
```

# Make list public
```sh
curl -XPUT 'http://localhost:1234/api/v1/share/readaccess' -u'iamchris:test123' -d '{
  "alist_uuid": "5d4c9869-1d26-567d-82be-497c3521368a",
  "action": "public"
}'
```

# View in learnalist-api server
```sh
http://localhost:1234/alists/5d4c9869-1d26-567d-82be-497c3521368a.html
```