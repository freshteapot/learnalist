CREATE TABLE IF NOT EXISTS alist_kv (
  uuid CHARACTER(36) not null primary key,

  list_type CHARACTER(3),

  body MEDIUMTEXT,
  user_uuid CHARACTER(36)

);
CREATE TABLE IF NOT EXISTS user (
  uuid CHARACTER(36) not null primary key,

  hash CHARACTER(20),

  username CHARACTER(36) NOT NULL UNIQUE
);
