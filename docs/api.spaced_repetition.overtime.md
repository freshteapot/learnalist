# Note
- User info on spaced repetition over time, only is cleared once all items have been added and have viewed once.

# Adding a list for spaced repetition overtime
- Codename: Dripfeed
- Dripfeed, adding an item overtime, tightly coupled to the user interacting with entries in spaced repetition

# Example api request
- Trigger adding the list items overtime

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/spaced-repetition/overtime' -d'
{
    "user_uuid":"user-123",
    "alist_uuid":"list-123"
}
'
```

# Full example
- Create user
- Login
- Create a list
- Add list for learning overtime

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

response=$(curl -XPOST \
-H "Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/alist' -d'
{
    "data": [
        "monday",
        "tuesday",
        "wednesday",
        "thursday",
        "friday",
        "saturday",
        "sunday"
    ],
    "info": {
        "title": "Days of the Week",
        "type": "v1"
    }
}
')
alistUUID=$(echo $response | jq -r '.uuid')

curl -XPOST \
-H "Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/spaced-repetition/overtime' -d@<(cat <<EOF
{
    "user_uuid": "${userUUID}",
    "alist_uuid": "${alistUUID}"
}
EOF
)
```

## Add list type v2

```sh
response=$(curl -XPOST \
-H "Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/alist' -d'
{
  "data": [
    {
      "from": "January",
      "to": "Januar"
    },
    {
      "from": "February",
      "to": "Februar"
    },
    {
      "from": "March",
      "to": "Mars"
    },
    {
      "from": "April",
      "to": "April"
    },
    {
      "from": "May",
      "to": "Mai"
    },
    {
      "from": "June",
      "to": "Juni"
    },
    {
      "from": "July",
      "to": "Juli"
    },
    {
      "from": "August",
      "to": "August"
    },
    {
      "from": "September",
      "to": "September"
    },
    {
      "from": "October",
      "to": "Oktober"
    },
    {
      "from": "November",
      "to": "November"
    },
    {
      "from": "December",
      "to": "Desember"
    }
  ],
  "info": {
    "title": "Months from English to Norwegian",
    "type": "v2",
    "labels": [
      "english",
      "norwegian"
    ]
  }
}
')
alistUUID=$(echo $response | jq -r '.uuid')
curl -XPOST \
-H "Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/spaced-repetition/overtime' -d@<(cat <<EOF
{
    "user_uuid": "${userUUID}",
    "alist_uuid": "${alistUUID}",
    "settings": {
        "show": "from"
    }
}
EOF
)
```

# Remove list from further adding overtime

```sh
curl -XDELETE \
-H "Authorization: Bearer ${token}" \
'http://127.0.0.1:1234/api/v1/spaced-repetition/overtime' -d@<(cat <<EOF
{
    "user_uuid": "${userUUID}",
    "alist_uuid": "${alistUUID}"
}
EOF
)
```

# Events
- event.ApiUserDelete
- event.CMDUserDelete
- event.SystemSpacedRepetition
- event.ApiSpacedRepetitionOvertime
- event.ApiSpacedRepetition
