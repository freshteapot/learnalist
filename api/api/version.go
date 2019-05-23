package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/api/version"
	"github.com/labstack/echo/v4"
)

type HttpGetVersionResponse struct {
	GitHash string `json:"gitHash"`
	GitDate string `json:"gitDate"`
	Version string `json:"version"`
	Url     string `json:"url"`
}

func (env *Env) V1GetVersion(c echo.Context) error {
	response := HttpGetVersionResponse{
		GitHash: version.GetGitHash(),
		GitDate: version.GetGitDate(),
		Version: version.GetVersion(),
		Url:     version.GetGitURL(),
	}

	return c.JSON(http.StatusOK, response)
}
