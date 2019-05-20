CREATE TABLE IF NOT EXISTS `casbin_rule` (
    `id` INTEGER,
    `p_type` VARCHAR(32) NOT NULL DEFAULT ''
    CHECK(
      typeof("p_type") = "text" AND
      length("p_type") <= 32
    ),
    `v0` VARCHAR(255) NOT NULL DEFAULT ''
    CHECK(
      typeof("v0") = "text" AND
      length("v0") <= 255
    ),
    `v1` VARCHAR(255) NOT NULL DEFAULT ''
    CHECK(
      typeof("v1") = "text" AND
      length("v1") <= 255
    ),
    `v2` VARCHAR(255) NOT NULL DEFAULT ''
    CHECK(
      typeof("v2") = "text" AND
      length("v2") <= 255
    ),
    `v3` VARCHAR(255) NOT NULL DEFAULT ''
    CHECK(
      typeof("v3") = "text" AND
      length("v3") <= 255
    ),
    `v4` VARCHAR(255) NOT NULL DEFAULT ''
    CHECK(
      typeof("v4") = "text" AND
      length("v4") <= 255
    ),
    `v5` VARCHAR(255) NOT NULL DEFAULT ''
    CHECK(
      typeof("v5") = "text" AND
      length("v5") <= 255
    ),
    PRIMARY KEY (`id`)
);
