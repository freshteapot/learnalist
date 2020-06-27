
# Add V1
```sh
curl -XPOST -H "Content-Type: application/json"  'http://localhost:1234/api/v1/spaced-repetition/' -u'iamchris:test123' -d '
{
  "show": "Hello",
  "data": "Hello",
  "kind": "v1"
}
'
```

# Add V2
```sh
curl -XPOST -H "Content-Type: application/json"  'http://localhost:1234/api/v1/spaced-repetition/' -u'iamchris:test123' -d '
{
  "show": "Mars",
  "data": {
    "from": "March",
    "to": "Mars"
  },
  "settings": {
    "show": "to"
  },
  "kind": "v2"
}
'
```


# Delete BY UUID
```sh
curl -XDELETE 'http://localhost:1234/api/v1/spaced-repetition/ba9277fc4c6190fb875ad8f9cee848dba699937f' -u'iamchris:test123'
```

# Get Next item to learn
```sh
curl -XGET 'http://localhost:1234/api/v1/spaced-repetition/next' -u'iamchris:test123'
```

# Get All
```sh
curl -XGET 'http://localhost:1234/api/v1/spaced-repetition/all' -u'iamchris:test123'
```


# Item was viewedx
```sh
curl -XPOST -H "Content-Type: application/json"  'http://localhost:1234/api/v1/spaced-repetition/viewed' -u'iamchris:test123' -d '
{
  "uuid": "75698c0f5a7b904f1799ceb68e2afe67ad987689",
  "action": "decr"
}
'
```
