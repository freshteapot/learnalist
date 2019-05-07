package models

type SimpleEvent struct {
	What     string
	WhatUuid string
	WhoUuid  string
	//Created  string
}

type AlistLabelLink struct {
	UserUuid  string
	AlistUuid string
}

type Label struct {
	Uuid     string
	Label    string
	UserUuid string
}
