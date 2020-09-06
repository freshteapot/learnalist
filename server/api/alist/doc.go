package alist

import "errors"

var (
	ErrorListFromValid             = errors.New("validate")
	ErrorSharingNotAllowedWithFrom = errors.New("sharing-not-allowed-with-from")
)
