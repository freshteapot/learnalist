


# Play along.
When the database is created, it is empty.

## You need a user first.
```sh
curl -XPOST 127.0.0.1:1234/v1/register -d'
{
    "username":"chris",
    "password":"test"
}
'
```

### Add a list of type v1.

```sh
curl -XPOST http://localhost:1234/v1/alist -u'chris:test' -d'
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
curl http://localhost:1234/v1/alist/by/me
```

### Add a list of type v2.

```sh
curl -XPOST http://localhost:1234/v1/alist -u'chris:test' -d'
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
curl -XPOST http://localhost:1234/v1/alist -u'chris:test' -d'
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
curl http://localhost:1234/v1/alist/by/me  -u'chris:test'
```

Filter based on list type v3.
```sh
curl http://localhost:1234/v1/alist/by/me?list_type=v3  -u'chris:test'
```

Or an individual list.
```sh
curl http://localhost:1234/v1/alist/{uuid} -u'chris:test'
```

#Create a list with a label or two
```sh
curl -s -w "%{http_code}\n" -XPOST http://localhost:1234/v1/alist -u'chris:test' -d'
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
curl -s -w "%{http_code}\n"  -XGET 'http://localhost:1234/v1/alist/by/me?labels=english' -u'chris:test'
```
