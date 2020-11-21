package mobile

var (
	EventMobileDeviceRegistered = "mobile.register"
)

type DeviceInfo struct {
	Token    string `json:"token"`
	UserUUID string `json:"user_uuid"`
}
