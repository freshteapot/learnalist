# App Settings api


## RemindV1

```sh
curl -i -XPUT \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/app-settings/remind_v1" -d'
{
    "spaced_repetition": {
        "push_enabled": 1
    }
}
'
```
