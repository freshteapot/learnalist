# Asset api

## How to upload an asset
- shared_with is optional = private, public
- shared_with defaults to private

### Request
```sh
curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
-F "file=@./server/e2e/testdata/sample.png" \
-F "shared_with=public" \
"http://localhost:1234/api/v1/assets/upload"
```

### Response
```sh
{
    "href":"/assets/affb846a-b02b-4d6e-a74b-afc19943dbf2/262fae17-b744-4987-b2dd-641ea1c7551a.png"
    "uuid": "262fae17-b744-4987-b2dd-641ea1c7551a",
    "ext": "png"
}
```

## Get asset
```sh
curl -XGET -I -H"Authorization: Bearer ${token}" \
'http://localhost:1234/assets/8f13afd7-71cf-4e20-843c-f45206fe85fd/78593152-037d-4732-a46b-8059273a2f27.png'
```

## Share asset
- private or public

```sh
curl -XPUT -H"Authorization: Bearer ${token}" \
'http://localhost:1234/api/v1/assets/share' -d'{
    "uuid": "509387be-ba8a-4b0f-991a-ef835ddd5c5d",
    "action": "private"
}'
```

## Delete asset
```sh
curl -XDELETE -I -H"Authorization: Bearer ${token}" \
'http://localhost:1234/api/v1/assets/78593152-037d-4732-a46b-8059273a2f27'
```
