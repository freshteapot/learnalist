CREATE TABLE IF NOT EXISTS spaced_repetition (
  uuid CHARACTER(36)  not null,
  body text,
  user_uuid CHARACTER(36),
  when_next DATETIME not null default (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  created DATETIME not null default (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE(user_uuid, uuid)
);

CREATE INDEX IF NOT EXISTS spaced_repetition_when_next ON spaced_repetition (user_uuid, when_next);
