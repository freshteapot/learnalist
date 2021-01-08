CREATE TABLE IF NOT EXISTS spaced_repetition_reminder (
  user_uuid CHARACTER(36) not null UNIQUE PRIMARY KEY,
  when_next DATETIME not null,
  last_active DATETIME not null,
  sent not null default 0
);

CREATE INDEX IF NOT EXISTS spaced_repetition_reminder_sent ON spaced_repetition_reminder (sent);
CREATE INDEX IF NOT EXISTS spaced_repetition_reminder_when_next ON spaced_repetition_reminder (when_next);
