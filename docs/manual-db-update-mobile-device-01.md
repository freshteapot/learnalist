# Update mobile device table to support app_identifier
- Installing from the start wont need these.

## Updating db/202011302208-mobile-device.sql
- Adding app_identifier and changing the index

## Upload new
- Upload new file to the server

```sql
ALTER TABLE mobile_device RENAME TO mobile_device_prev;
```

# Create
- take from ../server/db/202011302208-mobile-device.sql

# Insert

```sql
INSERT INTO
  mobile_device (user_uuid, app_identifier, token, created)
SELECT
  user_uuid, "plank:v1", token, created
FROM
  mobile_device_prev;
```

```sql
SELECT * FROM mobile_device;
```



```sql
DROP TABLE mobile_device_prev;
```
