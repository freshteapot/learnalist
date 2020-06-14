# Delete a user
- Only allowed to delete the user you are logged in with

## Request

```sh
curl -i -XDELETE -H"Authorization: Bearer ${token}" \
"http://localhost:1234/api/v1/user/${userUUID}"
```

## Response

```sh
{
  "message": "User has been removed"
}
```

# Sample script

- Login to get the access token and user uuid
- Delete the user

```sh
response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamchris",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')
curl -i -XDELETE -H"Authorization: Bearer ${token}" "http://localhost:1234/api/v1/user/${userUUID}"
```

