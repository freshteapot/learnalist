package api

import "github.com/freshteapot/learnalist-api/server/api/i18n"

type HTTPResponse struct {
	StatusCode int
	Body       []byte
}

type HTTPUserRegisterInput struct {
	Username string        `json:"username"`
	Password string        `json:"password"`
	Extra    HTTPUserExtra `json:"extra,omitempty"`
}

type HTTPUserExtra struct {
	DisplayName string `json:"display_name,omitempty"`
	CreatedVia  string `json:"created_via,omitempty"`
}

type HTTPUserRegisterResponse struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
}

type HTTPResponseMessage struct {
	Message string `json:"message"`
}

type HTTPLabelInput struct {
	Label string `json:"label"`
}

type HTTPGetVersionResponse struct {
	GitHash string `json:"gitHash"`
	GitDate string `json:"gitDate"`
	Version string `json:"version"`
	Url     string `json:"url"`
}

type HTTPShareListInput struct {
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

type HTTPShareListWithUserInput struct {
	UserUUID  string `json:"user_uuid"`
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

type HTTPLoginResponse struct {
	Token    string `json:"token"`
	UserUUID string `json:"user_uuid"`
}

type HTTPLoginRequest HTTPUserRegisterInput

type HTTPLogoutRequest struct {
	Kind     string `json:"kind"`
	UserUUID string `json:"user_uuid"`
	Token    string `json:"token"`
}

var (
	HTTPErrorResponse = HTTPResponseMessage{
		Message: i18n.InternalServerErrorFunny,
	}

	HTTPAccessDeniedResponse = HTTPResponseMessage{
		Message: i18n.AclHttpAccessDeny,
	}
)
