# Manual testing of endpoints

```sh
cd server/
rm ./server.db
ls db/*.sql | sort | xargs cat | sqlite3 ./server.db
go run commands/api/main.go --port=1234 --database=./server.db
```

# Add test user
```sh
curl -s -w "%{http_code}\n" -XPOST 127.0.0.1:1234/api/v1/register -d'
{
    "username":"iamchris",
    "password":"test123"
}
'
```

# Add a valid list v1
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": ["car"],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}'
```
Should return 200


# Try adding a list with an empty item.
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": [""],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}'
```
Should return
```sh
{"message":"Please refer to the documentation on list type v1"}
400
```

# Add a list with empty data
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}'
```
Should return 201


# Add a list with unknown type, should fail with 400.
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "na"
    }
}'
```
Should return 400


# Add a valid list v2
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": [{"from":"car", "to": "bil"}],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```
Should return 201.


# Add bad data and see a 400 response.
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": [{"from":"", "to": ""}],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```
Should return 400.
```
{"message":"Please refer to the documentation on list type v2"}
```


# Add list V2 with empty data
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```
Should return 201.

# Try putting a fake item.
(https://github.com/freshteapot/learnalist-api/issues/20)
```sh
curl -s -w "%{http_code}\n" -XPUT  http://127.0.0.1:1234/api/v1/alist/fakeuuid123 -u'iamchris:test123' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```
Should return 404
```
{"message":"List not found."}
```

# Delete a list that isnt in the system (https://github.com/freshteapot/learnalist-api/issues/21)
```sh
curl -s -w "%{http_code}\n" -XDELETE http://127.0.0.1:1234/api/v1/alist/fakeuuid123 -u'iamchris:test123'
```
Should return 404
```
{"message":"List not found."}
```

# Remove all lists via jq
```sh
curl -s  -XGET http://127.0.0.1:1234/api/v1/alist/by/me -u'iamchris:test123' | \
jq -r '.[] | .uuid' | \
awk '{cmd="curl -s -w \"%{http_code}\\n\" -XDELETE http://127.0.0.1:1234/api/v1/alist/"$1" -u'iamchris:test123'";print(cmd);system(cmd)}'
```

# Add a list with labels
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": ["car"],
    "info": {
        "title": "Days of the Week",
        "type": "v1",
        "labels": [
          "car",
          "water"
        ]
    }
}'
```
Should return 201


# Add a label
First time, it will return a 201
```sh
curl -s -w "%{http_code}\n"  -XPOST http://localhost:1234/api/v1/labels -uiamchris:test123 -d'
{
  "label": "water"
}'
```
Should return 201, and a list of current labels for the user.


Second time, it will return a 200
```sh
curl -s -w "%{http_code}\n"  -XPOST http://localhost:1234/api/v1/labels -uiamchris:test123 -d'
{
  "label": "water"
}'
```

# Get all labels from a user
```sh
curl -s -w "%{http_code}\n"  -XGET http://localhost:1234/api/v1/labels/by/me -u'iamchris:test123'
```

# Remove all labels
```sh
curl -s  -XGET http://127.0.0.1:1234/api/v1/labels/by/me -u'iamchris:test123' | \
jq -r '.[]' | \
awk '{cmd="curl -s -w \"%{http_code}\\n\" -XDELETE http://127.0.0.1:1234/api/v1/labels/"$1" -u'iamchris:test123'";print(cmd);system(cmd)}'
```
