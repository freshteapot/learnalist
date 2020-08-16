package models

const (
	SQL_INSERT_LIST                  = `INSERT INTO alist_kv(uuid, list_type, body, user_uuid) values(?, ?, ?, ?)`
	SQL_UPDATE_LIST                  = `UPDATE alist_kv SET list_type=?, body=?, user_uuid=? WHERE uuid=?`
	SQL_GET_ITEM_BY_UUID             = `SELECT uuid, body, user_uuid, list_type FROM alist_kv WHERE uuid = ?`
	SQL_DELETE_ITEM_BY_USER_AND_UUID = `
DELETE
FROM
	alist_kv
WHERE
	uuid=?
AND
	user_uuid=?
`
)

type SimpleEvent struct {
	What     string `db:"what"`
	WhatUuid string `db:"what_uuid"`
	WhoUuid  string `db:"who_uuid"`
	//Created  string
}

type AlistKV struct {
	Uuid     string `db:"uuid"`
	Body     string `db:"body"`
	UserUuid string `db:"user_uuid"`
	ListType string `db:"list_type"`
}

type GetListsByUserWithFiltersArgs struct {
	Labels   []string `db:"labels"`
	UserUuid string   `db:"user_uuid"`
	ListType string   `db:"list_type"`
}
