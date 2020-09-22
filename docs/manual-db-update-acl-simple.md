# Log of manual db updates Acl-Simple
- Installing from the start wont need these.

## Updating db/202006271346-spaced-repetition.sql
- Altering alist_uuid to be ext_uuid

# Alter table

```sql
ALTER TABLE acl_simple RENAME TO acl_simple_prev;
```

# Create
- take from ../server/db/201910050200-acl-simple.sql

# Insert

```sql
INSERT INTO
  acl_simple (ext_uuid, user_uuid, access)
SELECT
  alist_uuid, user_uuid, access
FROM
  acl_simple_prev;
```


```sql
DROP TABLE acl_simple_prev;
```
