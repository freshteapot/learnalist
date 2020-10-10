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
-H 'challenge: c:test:123' \
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
