# Challenge
## TODO
- [X] Store challenges
- [ ] Create UI for plank challenge
- [ ] Create UI for SRS challenge
- [ ] How to provide feedback for the challenges
- [ ] How to delete an entry via challenge = straight forward
- [ ] How to delete an entry via plank = has to listen
- [ ] Command to remove challenges that have no users in.
- [ ] Events for when record added to challenge
- [ ] Event for when user leaves a challenge
- [ ] Event for when user joins a challenge
- [ ] Event for when a challenge is created
- [ ] Create throw away accounts
- [ ] On logout on throw away account, send a delete message for the user?
- [ ] Delete user, does not delete challenge. Still not sure I want to.

```json
{
    "username": "iamtest",
    "password": "test123",
    "extra": {
        "displayName": "Chris",
        "created_via": "plank_v1",
        "secure": false,
        "created": "When RFC nano"
    }
}
```
- secure set internally based on header
- display limit to 20 characters
- created_via = reject if not valid
- user event process the info to build user preferences


## Development

# Create a challenge
## Login
```sh
response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')
```

## Create challenge
```sh
response=$(curl -XPOST \
-H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/challenge/' \
-d '
{
  "description": "hi",
  "kind": "plank-group"
}')
challengeUUID=$(echo $response | jq -r '.uuid')
```

## Get your challenges
- Either made
- Joined

```sh
curl \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/challenges/$userUUID"
```

## Join a challenge
### Register
```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
'
```

### Login
```sh
response=$(curl -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')
```

### Join challenge
```sh
curl -XPUT \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/challenge/$challengeUUID/join"
```

### View challenges
```sh
curl \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/challenges/$userUUID"
```

## Lookup challenge
```sh
curl \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/challenge/fcec0680-2aa2-4286-86f7-fbe3135722d8"
```


## Test
```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
'

response=$(curl -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')

response=$(curl -XPOST \
-H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/challenge/' \
-d '
{
  "description": "hi",
  "kind": "plank-group"
}')
challengeUUID=$(echo $response | jq -r '.uuid')

curl -XPOST \
-H "Authorization: Bearer ${token}" \
-H "x-challenge: ${challengeUUID}" \
'http://127.0.0.1:1234/api/v1/plank/' -d'
{
    "showIntervals": true,
    "intervalTime": 15,
    "beginningTime": 1602264153548,
    "currentTime": 1602264219291,
    "timerNow": 65743,
    "intervalTimerNow": 5681,
    "laps": 4
}
'
```
