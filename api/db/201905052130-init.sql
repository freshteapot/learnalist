PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS alist_kv (
  uuid CHARACTER(36)  not null primary key
  CHECK(
    typeof("uuid") = "text" AND
    length("uuid") <= 36
  ),
  list_type CHARACTER(3)
  CHECK(
    typeof("list_type") = "text" AND
    length("list_type") <= 3
  ),
  body text,
  user_uuid CHARACTER(36)
  CHECK(
    typeof("user_uuid") = "text" AND
    length("user_uuid") <= 36
  )
);
CREATE TABLE IF NOT EXISTS user (
  uuid CHARACTER(36) not null primary key,
  hash CHARACTER(20),
  username text NOT NULL UNIQUE
);
COMMIT;
