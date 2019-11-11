CREATE TABLE IF NOT EXISTS user_sessions (
  challenge TEXT NOT NULL UNIQUE PRIMARY KEY,
  token TEXT NOT NULL DEFAULT 'none',
  user_uuid TEXT NOT NULL DEFAULT 'none',
  created integer(4) not null default (strftime('%s','now'))
);
