# Get user info
```sh
curl -XGET \
-H "Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/user/info/${userUUID}"
```


{
  "user_uuid": "db6651ba-2a1f-49b7-9589-99d49816774f"
  "spaced_repetition": {
      "overtime": [
        "$alist_uuid"
      ]
  }
}
