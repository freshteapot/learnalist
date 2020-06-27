

```sh
curl -XPOST -H "Content-Type: application/json"  'http://localhost:1234/api/v1/spaced-repetition/' -u'iamchris:test123' -d '
{
  "show": "Hello",
  "data": "Hello",
  "kind": "v1"
}
'
```

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


curl -XDELETE 'http://localhost:1234/api/v1/spaced-repetition/75698c0f5a7b904f1799ceb68e2afe67ad987689' -u'iamchris:test123'


curl -XGET 'http://localhost:1234/api/v1/spaced-repetition/next' -u'iamchris:test123'
