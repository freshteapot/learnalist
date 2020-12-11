LAL_USERNAME="iamtest1"
LAL_PASSWORD="test123"
LAL_SERVER="http://localhost:1234"

CMD=$(
cat <<_EOF_
curl -XPOST '${LAL_SERVER}/api/v1/user/login' -d'
{
    "username":"${LAL_USERNAME}",
    "password":"${LAL_PASSWORD}"
}
'
_EOF_
)

response=$(eval $CMD)

userUUID=$(echo $response | jq -r '.user_uuid')
token=$(echo $response | jq -r '.token')

CMD=$(
cat <<_EOF_
curl -XPOST \
-H"Authorization: Bearer ${token}" \
-F "file=@./server/e2e/testdata/sample.png" \
-F "shared_with=public" \
"${LAL_SERVER}/api/v1/assets/upload"
_EOF_
)

response=$(eval $CMD)
assetUUID=$(echo $response | jq -r '.uuid')

CMD=$(cat <<_EOF_
curl -XPUT -H"Authorization: Bearer ${token}" \
'${LAL_SERVER}/api/v1/assets/share' -d'{
    "uuid": "${assetUUID}",
    "action": "private"
}'
_EOF_
)

response=$(eval $CMD)


CMD=$(cat <<_EOF_
curl -XDELETE -H"Authorization: Bearer ${token}" \
'${LAL_SERVER}/api/v1/assets/${assetUUID}'
_EOF_
)
