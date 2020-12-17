# Checking events in slack
- Delete user
- register
- post fake mobile token

```sh
response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')
curl -i -XDELETE -H"Authorization: Bearer ${token}" "http://localhost:1234/api/v1/user/${userUUID}"


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
  "token": "fake-token-123",
  "app_identifier": "remind:v1"
}')
```



```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamtest2",
    "password":"test123"
}
'

response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest2",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')
curl -i -XDELETE -H"Authorization: Bearer ${token}" "http://localhost:1234/api/v1/user/${userUUID}"
```
