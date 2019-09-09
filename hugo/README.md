cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.json data/lists/
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.md content/alists/



```sh
curl -XPUT 'http://localhost:1234/api/v1/share/readaccess' -u'iamchris:test123' -d '{
  "alist_uuid": "5d4c9869-1d26-567d-82be-497c3521368a",
  "action": "public"
}'
```

http://localhost:1234/alists/5d4c9869-1d26-567d-82be-497c3521368a.html
