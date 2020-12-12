# Mobile api

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


response=$(curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/mobile/register-device' \
-d '
{
  "token": "fake-token-123"
}')
```

## Add fake app

curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/mobile/register-device' \
-d '
{
  "token": "fake-token-123",
  "app_identifier": "fake-app"
}'


## Add new token

curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/mobile/register-device' \
-d '
{
  "token": "fake-token-123456",
  "app_identifier": "plank:v1"
}'

# Other valid option
curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/mobile/register-device' \
-d '
{
  "token": "fake-token-123456",
  "app_identifier": "remind:v1"
}'


