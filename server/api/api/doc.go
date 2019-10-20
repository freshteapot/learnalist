package api

import "github.com/freshteapot/learnalist-api/server/api/user"

type HttpUserRegisterInput user.RegisterInput

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
