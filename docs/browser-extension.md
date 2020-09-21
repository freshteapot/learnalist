```sh
curl -XPOST 'http://localhost:1234/api/v1/alist' -u'iamchris:test123' -d'
{
  "info": {
    "title": "Norwegian",
    "type": "v2",
    "from": {
      "kind": "quizlet",
      "ext_uuid": "test-123",
      "ref_url": "https://quizlet.com/test-123/norwegian/"
    }
  },
  "data": [
    {
      "from": "a",
      "to": "b"
    }
  ]
}
'
```
