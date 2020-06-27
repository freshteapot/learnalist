CREATE TABLE IF NOT EXISTS spaced_repetition (
  uuid CHARACTER(36)  not null primary key,
  body text,
  user_uuid CHARACTER(36),
  when_next integer(4) not null default (strftime('%s','now')),
  UNIQUE(user_uuid, uuid)
);
