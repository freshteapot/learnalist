# Api

| Method | Uri | Description | Status Codes |
| --- | --- | --- | --- |
| POST | /api/v1/spaced-repetition/ | Add entry for spaced based learning | TODO |
| DELETE | /api/v1/spaced-repetition/{UUID} | Delete entry | TODO |
| GET | /api/v1/spaced-repetition/next | Get next entry if ready | TODO |
| GET | /api/v1/spaced-repetition/all | Get next entry if ready | TODO |
| POST | /api/v1/spaced-repetition/viewed | Update entry to move forward or backwards thru the spaced based learning | TODO |


# Get Next item to learn
```sh
curl -XGET -u'iamchris:test123' \
'http://localhost:1234/api/v1/spaced-repetition/next'
```

# Get All entries
```sh
curl -XGET  -u'iamchris:test123' \
'http://localhost:1234/api/v1/spaced-repetition/all'
```



# Add Entry for learning
## V1
```sh
curl -XPOST -H "Content-Type: application/json" \
-u'iamchris:test123' \
'http://localhost:1234/api/v1/spaced-repetition/'  -d '
{
  "show": "Hello",
  "data": "Hello",
  "kind": "v1"
}
'
```

## Add V2
```sh
curl -XPOST -H "Content-Type: application/json" \
-u'iamchris:test123' \
'http://localhost:1234/api/v1/spaced-repetition/' -d '
{
  "show": "Mars",
  "data": {
    "from": "March",
    "to": "Mars"
  },
  "settings": {
    "show": "to"
  },
  "kind": "v2"
}
'
```


# Delete Entry by UUID
```sh
curl -XDELETE \
-u'iamchris:test123' \
'http://localhost:1234/api/v1/spaced-repetition/ba9277fc4c6190fb875ad8f9cee848dba699937f'
```


# Item was viewed
## Increase the gap till entry is seen again
```sh
curl -XPOST -H "Content-Type: application/json" \
-u'iamchris:test123'  \
'http://localhost:1234/api/v1/spaced-repetition/viewed' -d '
{
  "uuid": "75698c0f5a7b904f1799ceb68e2afe67ad987689",
  "action": "incr"
}
'
```


## Decrease the gap till entry is seen again
```sh
curl -XPOST -H "Content-Type: application/json" \
-u'iamchris:test123' \
'http://localhost:1234/api/v1/spaced-repetition/viewed' -d '
{
  "uuid": "75698c0f5a7b904f1799ceb68e2afe67ad987689",
  "action": "decr"
}
'
```


# Development
## Quickly reset all spaced repetitions
```sh
UPDATE spaced_repetition SET when_next=CURRENT_TIMESTAMP;
```
