# Log of manual db updates
- Installing from the start wont need these.

## Updating db/202006271346-spaced-repetition.sql
- Adding created column

# Alter table

```sql
ALTER TABLE spaced_repetition RENAME TO spaced_repetition_prev;
```

# Create
- take from db/202006271346-spaced-repetition.sql

# Insert

```sql
INSERT INTO
  spaced_repetition (uuid, body, user_uuid, when_next)
SELECT
  uuid, body, user_uuid, when_next
FROM
  spaced_repetition_prev;
```


```sql
DROP TABLE spaced_repetition_prev;
```

# Manually update the data

# Update the created first
```sql
UPDATE
    spaced_repetition
SET
    body=json_patch(body, '{"settings": {"created": "' || strftime('%Y-%m-%dT%H:%M:%SZ', created) || '" }}'),
    created=strftime('%Y-%m-%dT%H:%M:%SZ', created)
WHERE
    when_next NOT LIKE "%Z%";
```

# Update the when_next
```sql
UPDATE
    spaced_repetition
SET
    body=json_patch(body, '{"settings": {"when_next": "' || strftime('%Y-%m-%dT%H:%M:%SZ', when_next) || '" }}'),
    when_next=strftime('%Y-%m-%dT%H:%M:%SZ', when_next)
WHERE
    when_next NOT LIKE "%Z%";
```


# Diving into the data
```sql
SELECT
    strftime('%Y-%m-%dT%H:%M:%SZ', t.when_next),
    t.when_next,
    json_patch(body, '{"settings": {"when_next": "' || strftime('%Y-%m-%dT%H:%M:%SZ', t.when_next) || '" }}'),
    json_patch(body, '{"settings": {"created": "' || strftime('%Y-%m-%dT%H:%M:%SZ', t.created) || '" }}')
FROM
(
    SELECT *  FROM spaced_repetition ORDER BY when_next ASC LIMIT 10
) as t
WHERE
    t.when_next NOT LIKE "%Z%";
```
