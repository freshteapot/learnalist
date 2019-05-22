


# Play along.
When the database is created, it is empty.

## You need a user first.
```sh
curl -XPOST 'http://127.0.0.1:1234/v1/register' -d'
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
curl -XPOST 'http://localhost:1234/v1/alist' -u'iamchris:test123' -d'
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
curl 'http://localhost:1234/v1/alist/by/me'
```

### Add a list of type v2.

```sh
curl -XPOST 'http://localhost:1234/v1/alist' -u'iamchris:test123' -d'
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
curl -XPOST 'http://localhost:1234/v1/alist' -u'iamchris:test123' -d'
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
          "time": "1.46.4",
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
curl 'http://localhost:1234/v1/alist/by/me'  -u'iamchris:test123'
```

Filter based on list type v3.
```sh
curl 'http://localhost:1234/v1/alist/by/me?list_type=v3'  -u'iamchris:test123'
```

Or an individual list.
```sh
curl 'http://localhost:1234/v1/alist/{uuid}' -u'iamchris:test123'
```

#Create a list with a label or two
```sh
curl -s -w "%{http_code}\n" -XPOST http://localhost:1234/v1/alist -u'iamchris:test123' -d'
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
curl -s -w "%{http_code}\n"  -XGET 'http://localhost:1234/v1/alist/by/me?labels=english' -u'iamchris:test123'
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
Grant example
```sh
curl 'http://localhost:1234/v1/share/alist' -u'iamusera:test' -d '{
  "alist_uuid": "4e63bfd2-067f-5b58-8b8a-80a07f520825",
  "user_uuid": "5ce50aab-ae59-5a08-8483-5dabab92e563",
  "action": "grant"
}'
```

Revoke example
```sh
curl 'http://localhost:1234/v1/share/alist' -u'iamusera:test' -d '{
  "alist_uuid": "4e63bfd2-067f-5b58-8b8a-80a07f520825",
  "user_uuid": "5ce50aab-ae59-5a08-8483-5dabab92e563",
  "action": "revoke"
}'
```
