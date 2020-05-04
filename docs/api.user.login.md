# Login a user

## Request

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamchris",
    "password":"test123"
}
'
```

## Response

```
{
  "token": "7ab2d253-0c9f-46d6-a539-ca8b913aa480",
  "user_uuid": "731855f2-a70d-52f6-ada0-15a7690da0ea"
}
```
