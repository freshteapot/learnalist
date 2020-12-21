CREATE TABLE IF NOT EXISTS daily_reminder_settings (
  user_uuid CHARACTER(36) not null,
  app_identifier CHARACTER(36) not null,
  body text,
  when_next DATETIME not null,
  activity not null default 0,
  created DATETIME not null default (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
  UNIQUE(user_uuid, app_identifier)
);

CREATE INDEX IF NOT EXISTS daily_reminder_settings_when_next ON daily_reminder_settings (when_next);
CREATE INDEX IF NOT EXISTS daily_reminder_settings_user_uuid ON mobile_device (user_uuid);
