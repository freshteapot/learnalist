# Add asset for a user

## How to upload a file

```sh
curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
-F "file=@/Users/tinkerbell/git/learnalist-api/server/e2e/testdata/sample.png" \
"http://localhost:1234/api/v1/assets/upload"
```


{"href":"/assets/c98beb24-94f2-4858-bd64-0f193b0d7087/8fa20f32-6730-4d95-a03b-8422db0d180b.png"}


```sh
curl -i -XPOST \
-H"Authorization: Bearer ${token}" \
-H"Content-Type: multipart/form-data" \
"http://localhost:1234/api/v1/assets/upload"
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

