# Questions and Answers

## Via the api

## How can I get pretty print of the response?
Add the query string "pretty"
```sh
curl 'http://localhost:1234/api/v1/alist/by/me?pretty'
```

## Getting the response "Please refer to the documentation on list type v3"?
Check your data for the following rules

- Distance should not be empty.
- Stroke per minute (spm) should be between the range 10 and 50.
- Time should not be empty.
- Time is not valid format (xx:xx.xx).
- When should be YYYY-MM-DD.
- Per 500 (p500) should not be empty.
- Per 500 (p500) is not valid format (xx:xx.xx).

## How to link accounts from username and oauth2

- Manually for now, we will assume you want to copy the user uuid from "username and password" to "google".

### Get the current user id (from username and password)
```
SELECT
  uuid
FROM
  user
WHERE
  username=?
```

Record it = "XXX"

### Get the new user uuid, after you logged in for the first time.
```
SELECT
  user_uuid
FROM
  user_from_idp
WHERE
  idp="google"
AND
  identifier=?
```

Record it = "YYY"


### Update
- replace <XXX>
- replace <YYY>

```
UPDATE user_from_idp SET user_uuid="<XXX>" WHERE user_uuid="<YYY>";
```

```
UPDATE user_sessions SET user_uuid="<XXX>" WHERE user_uuid="<YYY>";
```

```
UPDATE oauth2_token_info SET user_uuid="<XXX>" WHERE user_uuid="<YYY>";
```

All done


# Testdata
```sh
cd ../hugo/
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.json data/lists/
cp testdata/5d4c9869-1d26-567d-82be-497c3521368a.md content/alists/
cd -
```

# Update the database with all changes.
```sh
ls server/db/*.sql | sort | xargs cat | sqlite3 server.db
```

# Update the database with a single file change.
```sh
cat  db/201905052144-labels.sql | sqlite3 test.db
```
