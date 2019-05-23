# Simple list (v1)

Data, provides an array of strings.

```json
[
  "monday",
  "tuesday",
  "wednesday",
  "thursday",
  "friday",
  "saturday",
  "sunday"
]
```

To create a list of type "v1", set type in the info object payload.

# Full example
```json
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
    "labels": []
  }
}
```

# Post it

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
    "type": "v1",
    "labels": []
  }
}
'
```

# Get your lists by filtering on Simplelist / v1
We add pretty to the query string, to return the json a little easier to read.
```sh
curl 'http://localhost:1234/v1/alist/by/me?list_type=v1&pretty'  -u'iamchris:test123'
```
or
```sh
curl 'http://localhost:1234/v1/alist/by/me?list_type=v1'  -u'iamchris:test123'
```
