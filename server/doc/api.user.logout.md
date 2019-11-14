# Logout a user

## Logout a single session for a user
### Request
```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/logout' -d'
{
  "kind": "token",
  "user_uuid":"731855f2-a70d-52f6-ada0-15a7690da0ea",
  "token":"7ab2d253-0c9f-46d6-a539-ca8b913aa480"
}
'
```

### Response
```
{
  "message": "Session 7ab2d253-0c9f-46d6-a539-ca8b913aa480, is now logged out"
}
```

## Logout all sessions for a user
### Request
```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/logout' -d'
{
  "kind": "user",
  "user_uuid":"731855f2-a70d-52f6-ada0-15a7690da0ea"
}
'
```

### Response
```
{
  "message": "All sessions have been logged out for user 731855f2-a70d-52f6-ada0-15a7690da0ea"
}
```
