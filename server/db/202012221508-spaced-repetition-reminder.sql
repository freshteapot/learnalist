CREATE TABLE IF NOT EXISTS spaced_repetition_reminder (
  user_uuid CHARACTER(36) not null UNIQUE PRIMARY KEY,
  when_next DATETIME not null,
  last_active DATETIME not null,
  sent not null default 0
);
