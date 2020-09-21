package assets

import (
	"errors"
	"time"

	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound = errors.New("not.found")
)

type AssetEntry struct {
	UUID      string    `db:"uuid"`
	UserUUID  string    `db:"user_uuid"`
	Extension string    `db:"extension"`
	Created   time.Time `db:"created"`
}

type AssetService struct {
	acl       acl.AclAsset
	directory string
	logEntry  *logrus.Entry
	repo      Repository
}

type Repository interface {
	SaveEntry(entry AssetEntry) error
	GetEntry(UUID string) (AssetEntry, error)
	DeleteEntry(userUUID string, UUID string) error
}

// HttpUploadResponse Response when asset uploaded
type HttpUploadResponse struct {
	Href string `json:"href"`
}
