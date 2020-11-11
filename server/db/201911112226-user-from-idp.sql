CREATE TABLE IF NOT EXISTS user_from_idp (
 user_uuid CHARACTER(36),
 idp TEXT NOT NULL,
 identifier TEXT NOT NULL,
 kind TEXT NOT NULL,
 info TEXT NOT NULL DEFAULT '',
 created integer(4) not null default (strftime('%s','now')),
 UNIQUE(user_uuid, idp, kind, identifier)
);
