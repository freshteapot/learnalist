CREATE TABLE IF NOT EXISTS challenge (
  uuid CHARACTER(36) not null primary key,
  body text,
  user_uuid CHARACTER(36),
  created DATETIME not null default (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE(user_uuid, uuid)
);

CREATE INDEX IF NOT EXISTS challenge_created ON challenge (user_uuid, created);

CREATE TABLE IF NOT EXISTS challenge_users (
  uuid CHARACTER(36) not null primary key,
  user_uuid CHARACTER(36) not null,
  info text,
  UNIQUE(uuid, user_uuid)
);

CREATE TABLE IF NOT EXISTS challenge_records (
  uuid CHARACTER(36) not null primary key,
  user_uuid CHARACTER(36) not null,
  ext_uuid CHARACTER(36) not null,
  UNIQUE(uuid, user_uuid, ext_uuid)
);
