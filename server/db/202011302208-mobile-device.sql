CREATE TABLE IF NOT EXISTS mobile_device (
  user_uuid CHARACTER(36) not null,
  token text,
  created DATETIME not null default (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS mobile_device_uniq ON mobile_device (user_uuid, token);
