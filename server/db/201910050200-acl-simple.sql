CREATE TABLE IF NOT EXISTS acl_simple (
 ext_uuid CHARACTER(36),
 user_uuid CHARACTER(36),
 access CHARACTER(100) not null,
 UNIQUE(ext_uuid, user_uuid, access)
);

CREATE INDEX IF NOT EXISTS ext_uuid_lookup ON acl_simple (ext_uuid);
