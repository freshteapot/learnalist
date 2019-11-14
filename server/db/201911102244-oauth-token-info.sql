CREATE TABLE IF NOT EXISTS oauth2_token_info (
 user_uuid CHARACTER(36) UNIQUE PRIMARY KEY,
 access_token TEXT NOT NULL,
 token_type TEXT NOT NULL,
 refresh_token TEXT NOT NULL,
 expiry integer(4) not null default (strftime('%s','now'))
);
