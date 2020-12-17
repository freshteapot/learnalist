package mobile

import "github.com/freshteapot/learnalist-api/server/pkg/openapi"

var (
	EventMobileDeviceRegistered = "mobile.register"
	EventMobileDeviceRemove     = "mobile.remove"
	EventMobileDeviceRemoved    = "mobile.removed"
)

type MobileRepository interface {
	SaveDeviceInfo(userUUID string, input openapi.HttpMobileRegisterInput) (int, error)
	DeleteByUser(userUUID string) error
	DeleteByToken(token string) error
}
type DeviceInfo struct {
	Token         string `json:"token"`
	UserUUID      string `json:"user_uuid"`
	AppIdentifier string `json:"app_identifier"`
}
