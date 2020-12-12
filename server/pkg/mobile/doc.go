package mobile

import "github.com/freshteapot/learnalist-api/server/pkg/openapi"

var (
	EventMobileDeviceRegistered = "mobile.register"
)

type MobileRepository interface {
	SaveDeviceInfo(userUUID string, input openapi.HttpMobileRegisterInput) (int, error)
	DeleteByUser(userUUID string) error
}
type DeviceInfo struct {
	Token    string `json:"token"`
	UserUUID string `json:"user_uuid"`
}
