# Manual testing of endpoints

```sh
cd api/
rm ./server.db
ls db/*.sql | sort | xargs cat | sqlite3 ./server.db
go run commands/api/main.go --port=1234 --database=./server.db
```

Add test user
```sh
curl -s -w "%{http_code}\n" -XPOST 127.0.0.1:1234/register -d'
{
    "username":"chris",
    "password":"test"
}
'
```

Add a valid list
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'chris:test' -d'
{
    "data": ["car"],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}'
```

Try adding a list with an empty item.
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'chris:test' -d'
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
{"message":"Failed to pass list type v1. Item cant be empty at position 0"}
400
```

Add a list with empty data
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'chris:test' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}'
```

Add a list with unknown type, should fail with 400.
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'chris:test' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "na"
    }
}'
```

Add a valid list v2
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'chris:test' -d'
{
    "data": [{"from":"car", "to": "bil"}],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```

Add bad data and see a 400 response.
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'chris:test' -d'
{
    "data": [{"from":"", "to": ""}],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```

Add list V2 with empty data, 200.
```sh
curl -s -w "%{http_code}\n" -XPOST  http://127.0.0.1:1234/alist -u'chris:test' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```

Try putting a fake item. (https://github.com/freshteapot/learnalist-api/issues/20)
```sh
curl -s -w "%{http_code}\n" -XPUT  http://127.0.0.1:1234/alist/fakeuuid123 -u'chris:test' -d'
{
    "data": [],
    "info": {
        "title": "Days of the Week",
        "type": "v2"
    }
}'
```

```sh
curl -s -w "%{http_code}\n" -XGET http://127.0.0.1:1234/alist/fakeuuid123 -u'chris:test'
```

Delete a list that isnt in the system (https://github.com/freshteapot/learnalist-api/issues/21)
```sh
curl -s -w "%{http_code}\n" -XDELETE http://127.0.0.1:1234/alist/fakeuuid123 -u'chris:test'
```

Remove all lists via jq
```sh
curl -s  -XGET http://127.0.0.1:1234/alist/by/me -u'chris:test' | \
jq -r '.[] | .uuid' | \
awk '{cmd="curl -s -w \"%{http_code}\\n\" -XDELETE http://127.0.0.1:1234/alist/"$1" -u'chris:test'";print(cmd);system(cmd)}'
```
