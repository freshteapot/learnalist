# Concept2 (v3)
This is a very niche list type.
The data within, is to log information from using the Concept2 indoor rowing machine.

```json
{
  "when": "2019-05-06", // {Year}-{Month}-{Day}
  "overall": {
    "time": "7:15.9", // M:S.MS
    "distance": 2000, // int
    "spm": 28, // int
    "p500": "1:48.9" // M:S.MS
  },
  "splits": [
    {
      "time": "1:46.4", // M:S.MS
      "distance": 500, // int
      "spm": 29, // int
      "p500": "1:58.0" // M:S.MS
    }
  ]
}
```

To create a list of type "v3", set type in the info object payload.

# Full example
```json
{
  "info": {
      "title": "A day on the rowing machine.",
      "type": "v3"
  },
  "data": [{
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
  }]
}
```

# Post it
This is an example of from and to, "from" English months "to" Norwegian.
```sh
curl -XPOST 'http://localhost:1234/v1/alist' -u'iamchris:test123' -d'
{
  "info": {
      "title": "A day on the rowing machine.",
      "type": "v3"
  },
  "data": [{
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
  }]
}
'
```

# Get your lists by filtering on Concept2 / v3
We add pretty to the query string, to return the json a little easier to read.
```sh
curl 'http://localhost:1234/v1/alist/by/me?list_type=v3&pretty'  -u'iamchris:test123'
```
or
```sh
curl 'http://localhost:1234/v1/alist/by/me?list_type=v3'  -u'iamchris:test123'
```
