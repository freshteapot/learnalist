package alist

import "errors"

var (
	ErrorListFromDomainMisMatch    = errors.New("domain-mis-match")
	ErrorListFromValid             = errors.New("validate")
	ErrorListFromKind              = errors.New("info.from.kind")
	ErrorSharingNotAllowedWithFrom = errors.New("sharing-not-allowed-with-from")
)
