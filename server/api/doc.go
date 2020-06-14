package api

type HttpUserRegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HttpUserRegisterResponse struct {
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
}
