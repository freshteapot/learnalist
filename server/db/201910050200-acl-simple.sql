CREATE TABLE IF NOT EXISTS acl_simple (
 alist_uuid CHARACTER(36),
 user_uuid CHARACTER(36),
 access CHARACTER(100) not null,
 UNIQUE(alist_uuid, user_uuid, access)
);

CREATE INDEX IF NOT EXISTS alist_uuid_lookup ON acl_simple (alist_uuid);
