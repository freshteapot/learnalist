# Log of manual db updates user-from-idp
- Installing from the start wont need these.

# Alter table

```sql
ALTER TABLE user_from_idp RENAME TO user_from_idp_prev;
```

# Create
- take from ../server/db/201910050200-acl-simple.sql


# Insert

```sql
INSERT INTO
  user_from_idp (user_uuid, idp, identifier, kind, info, created)
SELECT
  user_uuid, idp, identifier, kind, info, created
FROM
  user_from_idp_prev;
```


# Add
```sql
INSERT INTO
  user_from_idp (user_uuid, idp, identifier, kind, info, created)
SELECT
  user_uuid, idp, json_extract(info, "$.id") AS identifier, "id", info, created
FROM
  user_from_idp_prev;
```

# Remove the old entries
```sql
DELETE FROM user_from_idp WHERE kind="email";
```


# Run the site first ;)
```sql
DROP TABLE user_from_idp_prev;
```
