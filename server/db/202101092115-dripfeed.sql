CREATE TABLE IF NOT EXISTS dripfeed_item (
  dripfeed_uuid CHARACTER(36) not null,
  srs_uuid CHARACTER(36) not null,
  user_uuid CHARACTER(36) not null,
  alist_uuid CHARACTER(36) not null,
  body text not null,
  position integer(4) not null,
  UNIQUE(dripfeed_uuid, srs_uuid)
);

CREATE INDEX IF NOT EXISTS dripfeed_item_user ON dripfeed_item (user_uuid);
CREATE INDEX IF NOT EXISTS dripfeed_item_user_srs ON dripfeed_item (user_uuid, srs_uuid);
CREATE INDEX IF NOT EXISTS dripfeed_item_position ON dripfeed_item (dripfeed_uuid, position);
