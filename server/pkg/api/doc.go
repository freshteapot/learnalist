package api

type HttpResponse struct {
	StatusCode int
	Body       []byte
}

type HttpUserRegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HttpUserRegisterResponse struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
}

type HttpResponseMessage struct {
	Message string `json:"message"`
}

type HttpLabelInput struct {
	Label string `json:"label"`
}

type HttpGetVersionResponse struct {
	GitHash string `json:"gitHash"`
	GitDate string `json:"gitDate"`
	Version string `json:"version"`
	Url     string `json:"url"`
}

type HttpShareListInput struct {
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

type HttpShareListWithUserInput struct {
	UserUUID  string `json:"user_uuid"`
	AlistUUID string `json:"alist_uuid"`
	Action    string `json:"action"`
}

type HttpLoginResponse struct {
	Token    string `json:"token"`
	UserUUID string `json:"user_uuid"`
}
