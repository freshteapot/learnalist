CREATE INDEX IF NOT EXISTS alist_kv_user_uuid ON alist_kv (user_uuid);

CREATE TABLE IF NOT EXISTS label (
  uuid CHARACTER(36) not null primary key,
  label CHARACTER(20) not null,
  user_uuid CHARACTER(36) not null,
  UNIQUE(label, user_uuid)
);

CREATE TABLE IF NOT EXISTS alist_labels (
  alist_uuid CHARACTER(36) not null,
  label_uuid CHARACTER(36) not null,
  UNIQUE(alist_uuid, label_uuid)
);

CREATE TABLE IF NOT EXISTS simple_event (
  what CHARACTER(3) not null,
  what_uuid CHARACTER(36) not null,
  who_uuid CHARACTER(36) not null,
  created DATETIME NOT NULL DEFAULT (datetime(CURRENT_TIMESTAMP, 'utc'))
);

CREATE INDEX IF NOT EXISTS simple_event_who ON simple_event (who_uuid);
CREATE INDEX IF NOT EXISTS simple_event_who_what ON simple_event (who_uuid, what);


CREATE INDEX IF NOT EXISTS labels_by_user ON label (user_uuid);
