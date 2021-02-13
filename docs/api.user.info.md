# Get user info
```sh
curl -XGET \
-H "Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/user/info/${userUUID}"
```

# Response
```json
{
  "user_uuid": "a7dd5ba6-ba0c-4e5d-8b90-62a0e7e0ae36",
  "spaced_repetition": {
    "lists_overtime": [
      "223df1dc-c633-5f20-ae5d-e7b64ab22956"
    ]
  }
}
```

# Breakdown
## spaced_repetition.lists_overtime
Lists that are currently being added overitme for spaced repetition learning
