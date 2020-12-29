# Remind api

## Enable / Disable notifications in the remind app

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
'
response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')
```


```sh
curl -i -XPUT \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/remind/spaced-repetition/remind_v1" -d'
{
    "spaced_repetition": {
        "push_enabled": 0
    }
}
'
```

```sh
curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/mobile/register-device' \
-d '
{
  "token": "fake-token-123",
  "app_identifier": "remind_v1"
}'
```
