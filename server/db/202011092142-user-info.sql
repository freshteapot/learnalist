CREATE TABLE IF NOT EXISTS user_info (
  uuid CHARACTER(36) not null primary key,
  body text,
  UNIQUE(uuid)
);
