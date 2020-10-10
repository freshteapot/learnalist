CREATE TABLE IF NOT EXISTS plank (
  uuid CHARACTER(36)  not null primary key,
  body text,
  user_uuid CHARACTER(36),
  created DATETIME not null default (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE(user_uuid, uuid)
);

CREATE INDEX IF NOT EXISTS plank_user_created ON plank (user_uuid, created);
