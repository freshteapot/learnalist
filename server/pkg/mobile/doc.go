package mobile

import (
	"errors"

	"github.com/freshteapot/learnalist-api/server/pkg/openapi"
)

var (
	EventMobileDeviceRegistered = "mobile.register"
	EventMobileDeviceRemove     = "mobile.remove"
	EventMobileDeviceRemoved    = "mobile.removed"
)

type MobileRepository interface {
	SaveDeviceInfo(deviceInfo openapi.MobileDeviceInfo) (int, error)
	DeleteByUser(userUUID string) error
	DeleteByApp(userUUID string, appIdentifier string) error
	GetDeviceInfoByToken(token string) (openapi.MobileDeviceInfo, error)
}

var (
	ErrNotFound = errors.New("not.found")
)
