# Example

```sh
curl -XPOST 'http://127.0.0.1:1234/api/v1/dripfeed/' -d'
{
    "user_uuid":"user-123",
    "alist_uuid":"list-123"
}
'
```
# Test
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
'http://127.0.0.1:1234/api/v1/dripfeed/' -d@<(cat <<EOF
{
    "user_uuid": "${userUUID}",
    "alist_uuid": "${alistUUID}"
}
EOF
)
```


# Thoughts
- Table could be of all.
Ie

# Flow
event.ApiDripfeed/EventDripfeedInput -> add all events
Trigger GetNext event.SystemSpacedRepetition/spaced_repetition.EventSpacedRepetition

# Add all
# Exists =
# GetNext = lowest position next.
# Remove BY DripfeedUUID + SrUUID or position
# Body = Almost a complete SR but needs created updated

# dripfeedUUID = user_uuid/alist_uuid
CREATE TABLE IF NOT EXISTS dripfeed_item (
  dripfeed_uuid CHARACTER(36) not null primary key,
  sr_uuid CHARACTER(36),
  user_uuid CHARACTER(36),
  alist_uuid CHARACTER(36),
  body text not null,
  position integer(4) not null,
  UNIQUE(dripfeed_uuid, sr_uuid)
);

CREATE INDEX IF NOT EXISTS dripfeed_item_user ON dripfeed_item (user_uuid);
CREATE INDEX IF NOT EXISTS dripfeed_item_position ON dripfeed_item (dripfeed_uuid, position);
