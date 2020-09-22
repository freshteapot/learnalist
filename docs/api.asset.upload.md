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
curl -XGET -I -H"Authorization: Bearer ${token}" 'localhost:1234/assets/8f13afd7-71cf-4e20-843c-f45206fe85fd/78593152-037d-4732-a46b-8059273a2f27.png'
````

```sh
curl -XPUT -H"Authorization: Bearer ${token}" 'localhost:1234/api/v1/assets/share' -d'{
    "uuid": "78593152-037d-4732-a46b-8059273a2f27",
    "action": "private"
}'
```

```sh
curl -XDELETE -I -H"Authorization: Bearer ${token}" 'localhost:1234/api/v1/assets/78593152-037d-4732-a46b-8059273a2f27'
```
