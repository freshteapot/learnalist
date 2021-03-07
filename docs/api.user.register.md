# Register a user with a username and password



## Request

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
'
```

## Response
```
{
  "uuid": "0c6868e3-fc75-5161-be05-ce24ba59226e",
  "username": "iamtest1"
}
```


## Key lock down of this endpoint
- Use header "x-user-register"
- Key is set via server.userRegisterKey in yaml
- Overridden via env USER_REGISTER_KEY.

```sh
curl -i -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -H'x-user-register: hello1' -d'
{
    "username":"iamtest1",
    "password":"test123"
}
'
```


# Register with extra
- Give this user access to write public lists

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamtest1",
    "password":"test123",
    "extra": {
      "display_name": "Chris",
      "grant_public_list_write_access":"1"
    }
}
'
```
