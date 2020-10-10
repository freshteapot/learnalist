
# Poor mans import
lal_token=$lal_token
dev_token=$dev_token

CMD=$(
cat <<_EOF_
curl 'https://learnalist.net/api/v1/alist/by/me?labels=plank' \
  -H 'authorization: Bearer ${lal_token}' | jq -c '.[]|.data|.[]|fromjson'
_EOF_
)

PLANKS=$(eval $CMD)

for plank in $PLANKS; do
    # Not using challenge
    # -H 'challenge: c:test:123' \
    CMD=$(
cat <<_EOF_
curl -XPOST \
-H"Authorization: Bearer ${dev_token}" \
'http://127.0.0.1:1234/api/v1/plank/' -d'${plank}'
_EOF_
)
    response=$(eval $CMD)
    echo $response
done
