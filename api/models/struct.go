package models

const (
	ValidationWarningLabelToLong         = "The label cannot be longer than 20 characters."
	ValidationWarningLabelNotEmpty       = "The label cannot be empty."
	SuccessAlistNotFound                 = "List not found."
	InternalServerErrorMissingAlistUuid  = "Uuid is missing, possibly an internal error"
	InternalServerErrorMissingUserUuid   = "User.Uuid is missing, possibly an internal error"
	InternalServerErrorTalkingToDatabase = "Issue with talking to the database in %s."
)

type SimpleEvent struct {
	What     string `db:"what"`
	WhatUuid string `db:"what_uuid"`
	WhoUuid  string `db:"who_uuid"`
	//Created  string
}
