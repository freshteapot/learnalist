# Register
# Login
# Post list
# Add for over time
# Fake viewed
# View
# Check db

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
        "hello world",
        "hello Mars"
    ],
    "info": {
        "title": "Hello World",
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

# sqlite3 /tmp/learnalist/server.db .dump
# sqlite3 /tmp/learnalist/server.db "UPDATE spaced_repetition SET when_next=CURRENT_TIMESTAMP;"

curl -XPOST \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${token}" \
'http://localhost:1234/api/v1/spaced-repetition/viewed' -d '
{
  "uuid": "9c05511a31375a8a278a75207331bb1714e69dd1",
  "action": "incr"
}
'

# sqlite3 /tmp/learnalist/server.db .dump


curl -XGET \
-H "Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/user/info/${userUUID}" | jq




curl -XPOST \
-H "Content-Type: application/json" \
-H "Authorization: Bearer ${token}" \
'http://localhost:1234/api/v1/spaced-repetition/viewed' -d '
{
  "uuid": "fe5882814ce4cf1465e0e257a32dd24e2b532b73",
  "action": "incr"
}
'
