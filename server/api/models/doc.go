package models

import "github.com/freshteapot/learnalist-api/server/api/alist"

type Alist interface {
	Insert(aList alist.Alist) error
	Update(aList alist.Alist) error
	Remove(alistUUID string, userUUID string) error
	GetPublicLists() []alist.ShortInfo
}

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
