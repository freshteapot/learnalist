package mobile

var (
	EventMobileDeviceRegistered = "mobile.register"
)

type MobileRepository interface {
	SaveDeviceInfo(userUUID string, token string) (int, error)
	DeleteByUser(userUUID string) error
}
type DeviceInfo struct {
	Token    string `json:"token"`
	UserUUID string `json:"user_uuid"`
}
