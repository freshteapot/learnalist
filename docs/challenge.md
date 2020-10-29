# Challenge
## TODO
- [ ] Store challenges
- [ ] Create UI for plank challenge
- [ ] Create UI for SRS challenge
- [ ] How to provide feedback for the challenges
- [ ] How to delete an entry via challenge = straight forward
- [ ] How to delete an entry via plank = has to listen
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

```sh
curl -XGET -u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenges/633d1a70-bd5b-4ce6-afde-5495646d71e7'
```

```sh
curl -XPOST \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/' \
-d '
{
  "description": "hi",
  "kind": "plank-group"
}'
```

```sh
curl -XPUT \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/fake-123/join'
```

```sh
curl -XPUT \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/fake-123/leave'
```

```sh
curl -XGET \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/fake-123'
```


- Write the challenge
- Save the challenge as ChallengeEntry
- Look up challenges via acl?

description


338b8a8b-2edb-42d0-8dd3-40126487c73f


EXPLAIN SELECT
  *
FROM acl_simple
WHERE
  user_uuid="338b8a8b-2edb-42d0-8dd3-40126487c73f"
AND
  access LIKE "api:challenge:ef344ba6-0ea9-4049-b243-e354f39802b4:write:338b8a8b-2edb-42d0-8dd3-40126487c73f"



```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamchris",
    "password":"test123"
}
'
```

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamtest",
    "password":"test123"
}
'
```


curl -XPUT \
-u'iamtest:test123' \
'http://127.0.0.1:1234/api/v1/challenge/fake-123/join'






```sh
curl -XGET \
-u'iamchris:test123' \
"http://127.0.0.1:1234/api/v1/challenge/d41b4ac6-e402-423f-a769-48abc8818bd7"
```




response=$(curl -XPOST \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/' \
-d '
{
  "description": "hi",
  "kind": "plank-group"
}')

challengeUUID=$(echo $response | jq -r '.uuid')
token=$(echo $response | jq -r '.token')

curl -XGET \
-u'iamchris:test123' \
"http://127.0.0.1:1234/api/v1/challenge/$challengeUUID"


curl -XPUT \
-u'iamtest:test123' \
"http://127.0.0.1:1234/api/v1/challenge/$challengeUUID/join"


curl -XGET -u'iamtest:test123' \
'http://127.0.0.1:1234/api/v1/challenges/633d1a70-bd5b-4ce6-afde-5495646d71e7'


response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')

curl \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/challenges/$userUUID"


SELECT REPLACE(
  REPLACE("api:challenge:fake-123:write:27528ea5-f7c1-4a4a-a402-47c5716b9b2c", "api:challenge:", ""), ":write:27528ea5-f7c1-4a4a-a402-47c5716b9b2c", "")






response=$(curl -XPOST \
-u'iamchris:test123' \
'http://127.0.0.1:1234/api/v1/challenge/' \
-d '
{
  "description": "Daily plank",
  "kind": "plank-group"
}')

challengeUUID=$(echo $response | jq -r '.uuid')
token=$(echo $response | jq -r '.token')

curl -XGET \
-u'iamchris:test123' \
"http://127.0.0.1:1234/api/v1/challenge/$challengeUUID"
