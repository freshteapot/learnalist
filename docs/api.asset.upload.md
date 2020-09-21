# Add asset for a user

## How to upload an asset
// TODO add to openapi
- shared_with is optional = private, public
- shared_with defaults to private


// TODO change from my file
```sh
curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
-F "file=@/Users/tinkerbell/git/learnalist-api/server/e2e/testdata/sample.png" \
-F "shared_with=public" \
"http://localhost:1234/api/v1/assets/upload"
```


{"href":"/assets/affb846a-b02b-4d6e-a74b-afc19943dbf2/262fae17-b744-4987-b2dd-641ea1c7551a.png"}

```sh
curl -XGET -I -H"Authorization: Bearer ${token}" 'localhost:1234/assets/affb846a-b02b-4d6e-a74b-afc19943dbf2/262fae17-b744-4987-b2dd-641ea1c7551a.png'
````

```sh
curl -XPUT -I -H"Authorization: Bearer ${token}" 'localhost:1234/api/v1/assets/share' -d'{
    "uuid": "262fae17-b744-4987-b2dd-641ea1c7551a",
    "action": "private"
}'
```

```sh
curl -XDELETE -I -H"Authorization: Bearer ${token}" 'localhost:1234/api/v1/assets/6804625c-ff19-4cb8-aff8-faf7fc28582b'
```
