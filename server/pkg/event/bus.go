package event

import (
	"encoding/json"

	messagebus "github.com/vardius/message-bus"
)

const (
	ApiUserDelete             = "api.user.delete"
	ApiUserLogin              = "api.user.login"
	ApiUserRegister           = "api.user.register"
	TopicMonolog              = "lal.monolog"
	KindUserRegisterUsername  = "username"
	KindUserRegisterIDPGoogle = "idp:google"
)

func GetBus() messagebus.MessageBus {
	return bus
}

func SetBus(newBus messagebus.MessageBus) {
	bus = newBus
}

func EventLogToBytes(e Eventlog) []byte {
	b, _ := json.Marshal(e)
	return b
}
