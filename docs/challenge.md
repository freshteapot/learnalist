# Challenge
## TODO
- [ ] Store challenges
- [ ] Create UI for plank challenge
- [ ] Create UI for SRS challenge
- [ ] How to provide feedback for the challenges

## Development

### Send challenge with SRS

- include the challenge header

```sh
curl -XPOST -H "Content-Type: application/json" \
-u'iamchris:test123' \
-H 'x-challenge: c:test:123' \
'http://localhost:1234/api/v1/spaced-repetition/' -d '
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

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamchris",
    "password":"test123"
}
'
```

curl -XGET -u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenges/633d1a70-bd5b-4ce6-afde-5495646d71e7'

curl -XPOST \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/' \
-d '{"hi","you"}'

curl -XPUT \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/fake-123/join'

curl -XPUT \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/fake-123/leave'

curl -XGET \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/fake-123'
