package server

import (
	authenticateAlists "github.com/freshteapot/learnalist-api/server/alists/pkg/authenticate"
	alists "github.com/freshteapot/learnalist-api/server/alists/server"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
)

func InitAlists(
	manager *alists.Manager,
	userSession user.Session,
	userWithUsernameAndPassword user.UserWithUsernameAndPassword,
	pathToPublicDirectory string,
) {

	authConfig := authenticate.Config{
		LookupBasic:  userWithUsernameAndPassword.Lookup,
		LookupBearer: userSession.GetUserUUIDByToken,
		Skip:         authenticateAlists.SkipAuth,
	}

	server.GET("/logout.html", manager.Logout)
	server.GET("/lists-by-me.html", manager.GetMyLists, authenticate.Auth(authConfig))
	server.GET("/alistsbyuser/:uuid.html", manager.GetMyListsByURI, authenticate.Auth(authConfig))

	alists := server.Group("/alist")
	alists.Use(authenticate.Auth(authConfig))

	alists.GET("/*", manager.GetAlist)
	server.Static("/", pathToPublicDirectory)
}
