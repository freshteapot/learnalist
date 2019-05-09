package models

const (
	ValidationWarningLabelToLong = "The label cannot be longer than 20 characters."
)

type SimpleEvent struct {
	What     string `db:"what"`
	WhatUuid string `db:"what_uuid"`
	WhoUuid  string `db:"who_uuid"`
	//Created  string
}
