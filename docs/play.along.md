# Play along.

```sh
make clear-site
make rebuild-db
make develop
```

When the database is created, it is empty.

## You need a user first.
```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamchris",
    "password":"test123"
}
'
```
Response is
```sh
{"uuid":"1ff95121-1570-5e96-8bd9-bb62fac0b999","username":"iamchris"}
```

### Add a list of type v1.

```sh
curl -XPOST 'http://localhost:1234/api/v1/alist' -u'iamchris:test123' -d'
{
    "data": [
        "monday",
        "tuesday",
        "wednesday",
        "thursday",
        "friday",
        "saturday",
        "sunday"
    ],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}
'
```

### Get all lists created by you.
```sh
curl 'http://localhost:1234/api/v1/alist/by/me'
```

```sh
curl 'http://localhost:1234/api/v1/alist/by/me' -u'iamchris:test123'
```

### Add a list of type v2.

```sh
curl -XPOST 'http://localhost:1234/api/v1/alist' -u'iamchris:test123' -d'
{
    "data": [
        {
            "from": "chris",
            "to": "christopher"
        }
    ],
    "info": {
        "title": "A list of key:value pairs.",
        "type": "v2"
    }
}'
```

### Add a list of type v3 (concept2)
```sh
curl -XPOST 'http://localhost:1234/api/v1/alist' -u'iamchris:test123' -d'
{
  "data": [
    {
      "when": "2019-05-06",
      "overall": {
        "time": "7:15.9",
        "distance": 2000,
        "spm": 28,
        "p500": "1:48.9"
      },
      "splits": [
        {
          "time": "1:46.4",
          "distance": 500,
          "spm": 29,
          "p500": "1:58.0"
        }
      ]
    }
  ],
  "info": {
      "title": "A list to record rows.",
      "type": "v3"
  }
}'
```

Again, query all the lists by you.
```sh
curl 'http://localhost:1234/api/v1/alist/by/me'  -u'iamchris:test123'
```

Filter based on list type v3.
```sh
curl 'http://localhost:1234/api/v1/alist/by/me?list_type=v3'  -u'iamchris:test123'
```

Or an individual list.
```sh
curl 'http://localhost:1234/api/v1/alist/{uuid}' -u'iamchris:test123'
```

#Create a list with a label or two
```sh
curl -s -w "%{http_code}\n" -XPOST http://localhost:1234/api/v1/alist -u'iamchris:test123' -d'
{
    "data": [
        "monday",
        "tuesday",
        "wednesday",
        "thursday",
        "friday",
        "saturday",
        "sunday"
    ],
    "info": {
        "title": "Days of the Week",
        "type": "v1",
        "labels": [
          "english"
        ]
    }
}
'
```

Now try querying for this list via the labels filter
```sh
curl -s -w "%{http_code}\n"  -XGET 'http://localhost:1234/api/v1/alist/by/me?labels=english' -u'iamchris:test123'
```


# Share a list
- Supports two actions (grant and revoke).

```go
type HttpShareListWithUserInput struct {
	UserUUID  string `json:"user_uuid"`
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}
```

Create new user
```sh
curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamusera",
    "password":"test123"
}
' | jq '.uuid'
```

```sh
curl -XPOST 'http://localhost:1234/api/v1/alist' -u'iamchris:test123' -d'
{
    "data": [
        "monday",
        "tuesday",
        "wednesday",
        "thursday",
        "friday",
        "saturday",
        "sunday"
    ],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}
' | jq '.uuid'
```


Grant example

```sh
curl -XPUT 'http://localhost:1234/api/v1/share/readaccess' -u'iamchris:test123' -d '{
  "alist_uuid": "14ae1d04-f26a-524c-8539-2a7059f359e8",
  "user_uuid": "f3572f35-eb7d-5a16-a13f-925e3dd270f6",
  "action": "grant"
}'
```

Revoke example
```sh
curl -XPUT 'http://localhost:1234/api/v1/share/readaccess' -u'iamchris:test123' -d '{
  "alist_uuid": "14ae1d04-f26a-524c-8539-2a7059f359e8",
  "user_uuid": "f3572f35-eb7d-5a16-a13f-925e3dd270f6",
  "action": "revoke"
}'
```


Share list with the public
```sh
curl -XPUT 'http://localhost:1234/api/v1/share/alist' -u'iamchris:test123' -d '{
  "alist_uuid": "14ae1d04-f26a-524c-8539-2a7059f359e8",
  "action": "public"
}'
```

curl -XGET 'http://localhost:1234/api/v1/alist/14ae1d04-f26a-524c-8539-2a7059f359e8' -u'iamusera:test123'

Share list with friends
```sh
curl -XPUT 'http://localhost:1234/api/v1/share/alist' -u'iamchris:test123' -d '{
  "alist_uuid": "14ae1d04-f26a-524c-8539-2a7059f359e8",
  "action": "friends"
}'
```

Share list only with owner
```sh
curl -XPUT 'http://localhost:1234/api/v1/share/alist' -u'iamchris:test123' -d '{
  "alist_uuid": "14ae1d04-f26a-524c-8539-2a7059f359e8",
  "action": "private"
}'
```
