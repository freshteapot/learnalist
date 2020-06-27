

curl -XPOST -H "Content-Type: application/json"  'http://localhost:1234/api/v1/spaced-repetition' -u'iamchris:test123' -d '
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


curl -XDELETE 'http://localhost:1234/api/v1/spaced-repetition' -u'iamchris:test123' -d '
{
  "uuid": "Mars",
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
