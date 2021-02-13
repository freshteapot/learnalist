package dripfeed

import (
	"github.com/freshteapot/learnalist-api/server/api/alist"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/sirupsen/logrus"
)

type DripfeedRepository interface {
	AddAll(dripfeedUUID string, userUUID string, alistUUID string, items []string) error
	DeleteByUUIDAndUserUUID(dripfeedUUID string, userUUID string) error
	DeleteByUser(userUUID string) error
	DeleteAllByUserUUIDAndSpacedRepetitionUUID(userUUID string, srsUUID string) error
	Exists(dripfeedUUID string) (bool, error)
	// GetNext return the next spaced entry (v1 or v2)
	GetNext(dripfeedUUID string) (RepoItem, error)
	GetInfo(dripfeedUUID string) (openapi.SpacedRepetitionOvertimeInfo, error)
	SaveInfo(input openapi.SpacedRepetitionOvertimeInfo) error
}

type DripfeedService struct {
	repo       DripfeedRepository
	aclRepo    acl.AclReaderList
	listRepo   alist.DatastoreAlists
	logContext logrus.FieldLogger
}

// Nice for getting things out, not nice for rebuilding
type SpacedRepetitionSettingsExtID struct {
	Settings struct {
		ExtID string `json:"ext_id"`
	} `json:"settings"`
}

type SpacedRepetitionSettingsBase struct {
	Settings spaced_repetition.HTTPRequestInputSettings `json:"settings"`
}

type SpacedRepetitionUUID struct {
	UUID string `json:"uuid"`
}

type EventDripfeedDelete struct {
	UserUUID     string `json:"user_uuid"`
	DripfeedUUID string `json:"dripfeed_uuid"`
}

type EventDripfeedInputBase struct {
	UserUUID  string `json:"user_uuid"`
	AlistUUID string `json:"alist_uuid"`
	Kind      string `json:"kind"` // This is the list_type, at some point I will drop list_type :P
}

type EventDripfeedInputInfo struct {
	Info EventDripfeedInputBase `json:"info"`
}

type EventDripfeedInputV1 struct {
	Info EventDripfeedInputBase `json:"info"`
	Data alist.TypeV1           `json:"data"`
}

type EventDripfeedInputV2 struct {
	Info     EventDripfeedInputBase                               `json:"info"`
	Settings openapi.SpacedRepetitionOvertimeInputV2AllOfSettings `json:"settings"`
	Data     alist.TypeV2                                         `json:"data"`
}

// Used to map from db to spacedRepetition
type RepoItem struct {
	SrsUUID      string
	SrsKind      string
	SrsBody      []byte
	Position     int
	DripfeedUUID string
	UserUUID     string
	AlistUUID    string
}

var (
	EventDripfeedAdded    = "dripfeed.added"
	EventDripfeedRemoved  = "dripfeed.removed"
	EventDripfeedFinished = "dripfeed.finished"
)

// openapi.SpacedRepetitionOvertimeInfo
