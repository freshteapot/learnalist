# Plank api

## Get the history
```sh
curl -H"Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/plank/history'
```

## Add entry
```json
curl -XPOST 'http://127.0.0.1:1234/api/v1/plank' -d'
{
    "showIntervals": true,
    "intervalTime": 15,
    "beginningTime": 1602264153548,
    "currentTime": 1602264219291,
    "timerNow": 65743,
    "intervalTimerNow": 5681,
    "laps": 4
}
'
```

# Do it yourself

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/user/register' -d'
{
    "username":"iamchris",
    "password":"test123"
}
'
```

```sh
response=$(curl -s -XPOST 'http://127.0.0.1:1234/api/v1/user/login' -d'
{
    "username":"iamchris",
    "password":"test123"
}
')
userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')
```

```sh
curl -XPOST \
-H"Authorization: Bearer ${token}" \
-H 'challenge: c:test:123' \
'http://127.0.0.1:1234/api/v1/plank/' -d'
{
    "showIntervals": true,
    "intervalTime": 15,
    "beginningTime": 1602264153548,
    "currentTime": 1602264219291,
    "timerNow": 65743,
    "intervalTimerNow": 5681,
    "laps": 4
}
'
```

# Delete Entry by UUID
```sh
curl -XDELETE \
-u'iamchris:test123' \
'http://localhost:1234/api/v1/plank/ba9277fc4c6190fb875ad8f9cee848dba699937f'
```
