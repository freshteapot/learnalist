CREATE INDEX IF NOT EXISTS alist_kv_user_uuid ON alist_kv (user_uuid);

CREATE TABLE IF NOT EXISTS user_labels (
  label TEXT not null
  CHECK(
    typeof("label") = "text" AND
    length("label") <= 20
  ),
  user_uuid CHARACTER(36) not null,
  UNIQUE(user_uuid, label)
);

CREATE TABLE IF NOT EXISTS alist_labels (
  alist_uuid CHARACTER(36) not null,
  user_uuid CHARACTER(36) not null,
  label CHARACTER(20) not null
  CHECK(
    typeof("label") = "text" AND
    length("label") <= 20
  ),
  UNIQUE(alist_uuid, user_uuid, label)
);


CREATE INDEX IF NOT EXISTS labels_for_user ON user_labels (user_uuid);
CREATE INDEX IF NOT EXISTS labels_for_alist ON alist_labels (alist_uuid);
CREATE INDEX IF NOT EXISTS labels_for_alist_by_user ON alist_labels (user_uuid);
CREATE INDEX IF NOT EXISTS labels_for_alist_by_label ON alist_labels (label);
