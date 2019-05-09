


# Play along.
When the database is created, it is empty.

## You need a user first.
```sh
curl -XPOST 127.0.0.1:1234/register -d'
{
    "username":"chris",
    "password":"test"
}
'
```

### Add a list of type v1.

```sh
curl -XPOST http://localhost:1234/alist -u'chris:test' -d'
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
curl http://localhost:1234/alist/by/me
```

### Add a list of type v2.

```sh
curl -XPOST http://localhost:1234/alist -uchris:test -d'
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

Again, query all the lists by you.
```sh
curl http://localhost:1234/alist/by/me
```

Or an individual list.
```sh
curl http://localhost:1234/alist/{uuid}
```

#Create a list with a label or two
```sh
curl -s -w "%{http_code}\n" -XPOST http://localhost:1234/alist -u'chris:test' -d'
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
curl -s -w "%{http_code}\n"  -XGET 'http://localhost:1234/alist/by/me?labels=english' -u'chris:test'
```
