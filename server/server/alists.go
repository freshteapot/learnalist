package server

import (
	authenticateAlists "github.com/freshteapot/learnalist-api/server/alists/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"

	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
)

func customHTTPErrorHandler(err error, c echo.Context) {
	// TODO custom for list
	// TODO custom for not lists
	// TODO this is global :(
	// Can know if its application/json
	// Can know if it is api/
	// Can know if it is alists/
	// Can know if logged in

	/*
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
		}
		//errorPage := fmt.Sprintf("%d.html", code)
	*/
	log.Info("HELLO")
	log.Info(c.Request().URL.Path)
	errorPage := "/tmp/learnalist-api/site-cache/alist/404.html"
	if err := c.File(errorPage); err != nil {
		c.Logger().Error(err)
	}
	c.Logger().Error(err)
}

func InitAlists(acl acl.Acl, dal models.Datastore, hugoHelper *hugo.HugoHelper) {
	m := alists.Manager{
		Acl:             acl,
		Datastore:       dal,
		SiteCacheFolder: config.SiteCacheFolder,
		HugoHelper:      *hugoHelper,
	}

	authConfig := authenticate.Config{
		LookupBasic:  m.Datastore.UserWithUsernameAndPassword().Lookup,
		LookupBearer: m.Datastore.UserSession().GetUserUUIDByToken,
		Skip:         authenticateAlists.SkipAuth,
	}

	alists := server.Group("/alist")
	alists.Use(authenticate.Auth(authConfig))

	alists.GET("/*", m.GetAlist)

	// TODO http://localhost:1234/lists-by-me.html
	// TODO block access to the user files (alistsbyuser)
	server.Static("/", config.SiteCacheFolder)
	server.HTTPErrorHandler = customHTTPErrorHandler
}
