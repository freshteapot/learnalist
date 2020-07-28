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
