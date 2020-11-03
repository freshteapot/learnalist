# Challenge
## TODO
- [X] Store challenges
- [ ] Create UI for plank challenge
- [ ] Create UI for SRS challenge
- [ ] How to provide feedback for the challenges
- [ ] How to delete an entry via challenge = straight forward
- [ ] How to delete an entry via plank = has to listen
## Development

# Create a challenge
## Login
```sh
response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest",
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


