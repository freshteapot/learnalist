package assets

import (
	"time"

	"github.com/sirupsen/logrus"
)

type AssetEntry struct {
	UUID      string    `db:"uuid"`
	UserUUID  string    `db:"user_uuid"`
	Extension string    `db:"extension"`
	Created   time.Time `db:"created"`
}

type AssetService struct {
	directory string
	logEntry  *logrus.Entry
	repo      Repository
}

type Repository interface {
	SaveEntry(entry AssetEntry) error
}

// HttpUploadResponse Response when asset uploaded
type HttpUploadResponse struct {
	Href string `json:"href"`
}
