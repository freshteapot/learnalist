package models

type SimpleEvent struct {
	What     string `db:"what"`
	WhatUuid string `db:"what_uuid"`
	WhoUuid  string `db:"who_uuid"`
	//Created  string
}

type AlistLabelLink struct {
	LabelUuid string `db:"label_uuid"`
	AlistUuid string `db:"alist_uuid"`
}

type Label struct {
	Uuid      string `json:"uuid" db:"uuid"`
	Label     string `json:"label" db:"label"`
	UserUuid  string `json:"-" db:"user_uuid"`
	AlistUuid string `json:"-" db:"alist_uuid"`
}
