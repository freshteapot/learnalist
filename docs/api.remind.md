# Remind
- Able to remind daily at a particular time
- Currently supports "push".
- Only works with "remind_v1" app.
- TZ is based on IANA Time Zone database.

# Get all daily reminders
Look at [user info](./api.user.info.md).

# Get my daily reminders for specific app
- GET /api/v1/remind/daily/:appIdentifier
```sh
curl -H"Authorization: Bearer ${token}" "http://127.0.0.1:1234/api/v1/remind/daily/remind_v1"
```

## Send me a push notification at 9:00am every day in the timezone Oslo
- PUT /api/v1/remind/daily/
```sh
curl -i -XPUT \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/remind/daily/" -d'
{
    "time_of_day": "09:00",
    "tz": "Europe/Oslo",
    "medium": ["push"],
    "app_identifier": "remind_v1"
}
'
```

# Delete daily reminders for specific app
- DELETE /api/v1/remind/daily/:appIdentifier
```sh
curl -XDELETE \
-H"Authorization: Bearer ${token}" \
"http://127.0.0.1:1234/api/v1/remind/daily/remind_v1"
```
