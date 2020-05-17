CREATE TABLE IF NOT EXISTS fixup_history (
 the_fix CHARACTER(36) UNIQUE PRIMARY KEY,
 created integer(4) not null default (strftime('%s','now'))
);
