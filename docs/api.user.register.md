# Register a user with a username and password

## Request

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamchris",
    "password":"test123"
}
'
```

## Response
```
{
  "uuid": "0c6868e3-fc75-5161-be05-ce24ba59226e",
  "username": "iamchris"
}
```
