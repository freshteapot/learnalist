package api

import (
	"net/http"

	"github.com/freshteapot/learnalist-api/server/api/version"
	"github.com/freshteapot/learnalist-api/server/pkg/api"
	"github.com/labstack/echo/v4"
)

func (m *Manager) V1GetVersion(c echo.Context) error {
	response := api.HttpGetVersionResponse{
		GitHash: version.GetGitHash(),
		GitDate: version.GetGitDate(),
		Version: version.GetVersion(),
		Url:     version.GetGitURL(),
	}

	return c.JSON(http.StatusOK, response)
}
