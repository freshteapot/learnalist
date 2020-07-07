CREATE TABLE IF NOT EXISTS user_assets (
  uuid CHARACTER(36)  not null primary key,
  user_uuid CHARACTER(36),
  extension CHARACTER(36),
  created DATETIME not null default (strftime('%s','now'))
);
