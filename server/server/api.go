package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/freshteapot/learnalist-api/server/alists/pkg/hugo"
	"github.com/freshteapot/learnalist-api/server/api/api"
	authenticateApi "github.com/freshteapot/learnalist-api/server/api/authenticate"
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/api/uuid"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/authenticate"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	userSqlite "github.com/freshteapot/learnalist-api/server/pkg/user/sqlite"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"

	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
)

func InitApi(db *sqlx.DB, acl acl.Acl, dal *models.DAL, hugoHelper hugo.HugoHelper, oauthHandlers *oauth.Handlers, logger *logrus.Logger) {

	userManagement := user.NewManagement(
		userSqlite.NewSqliteManagementStorage(db),
		hugoHelper,
		event.NewInsights(logger),
	)

	m := api.NewManager(dal, userManagement, acl, "", hugoHelper, *oauthHandlers, logger)

	authConfig := authenticate.Config{
		LookupBasic:  m.Datastore.UserWithUsernameAndPassword().Lookup,
		LookupBearer: m.Datastore.UserSession().GetUserUUIDByToken,
		Skip:         authenticateApi.Skip,
	}

	v1 := server.Group("/api/v1")

	v1.Use(authenticate.Auth(authConfig))

	v1.GET("/version", m.V1GetVersion)

	v1.POST("/user/register", m.V1PostRegister)
	v1.POST("/user/login", m.V1PostLogin)
	v1.POST("/user/logout", m.V1PostLogout)
	v1.DELETE("/user/:uuid", m.V1DeleteUser)

	// Route => handler
	v1.GET("/", m.V1GetRoot)
	v1.GET("/alist/:uuid", m.V1GetListByUUID)
	v1.GET("/alist/by/me", m.V1GetListsByMe)

	//e.POST("/alist/v1", m.V1PostAlist)
	//e.POST("/alist/v2", m.V1PostAlist)
	//e.POST("/alist/v3", m.V1PostAlist)
	//e.POST("/alist/v4", m.V1PostAlist)
	v1.POST("/alist", m.V1SaveAlist)
	v1.PUT("/share/alist", m.V1ShareAlist)
	v1.PUT("/share/readaccess", m.V1ShareListReadAccess)
	v1.PUT("/alist/:uuid", m.V1SaveAlist)
	v1.DELETE("/alist/:uuid", m.V1RemoveAlist)
	// Labels
	v1.POST("/labels", m.V1PostUserLabel)
	v1.GET("/labels/by/me", m.V1GetUserLabels)
	v1.DELETE("/labels/:label", m.V1RemoveUserLabel)

	// Oauth
	v1.GET("/oauth/google/redirect", m.V1OauthGoogleRedirect)
	v1.GET("/oauth/google/callback", m.V1OauthGoogleCallback)

	srs := server.Group("/api/v1/spaced-repetition")
	srs.Use(authenticate.Auth(authConfig))
	sseServer := sse.New()

	sseServer.AutoReplay = false
	srsServer := spaced_repetition.NewService(db)
	srsServer.Endpoints(srs)
	srs.GET("/events", sseForEcho(sseServer), echoExampleMiddleware, echo.WrapMiddleware(exampleMiddleware))

	// Lets come back to server side events and fake it for now
	//events := make(chan *sse.Event)
	// client := sse.NewClient("/api/v1/spaced-repetition/events")

	// TODO change to 1m
	// TODO how to get active users
	// Possible help https://github.com/ReneKroon/ttlcache
	//_cron.AddFunc("@every 1s", srsServer.CheckForNewItems)

	streamKey := "hello"
	// I wonder how light this is?
	sseServer.CreateStream(streamKey)

	for index, _ := range sseServer.Streams {
		fmt.Println(index)
	}

	type event struct {
		UUID string    `json:"uuid"`
		When time.Time `json:"when"`
	}

	now := time.Now()
	e := event{UUID: "tine", When: now}
	b, _ := json.Marshal(e)
	sseServer.Publish(streamKey, &sse.Event{
		Data: b,
	})

	//time.Sleep(2000 * time.Millisecond)

	e = event{UUID: "tine1", When: now}
	b, _ = json.Marshal(e)
	sseServer.Publish(streamKey, &sse.Event{
		Data: b,
	})
}

func sseForEcho(server *sse.Server) echo.HandlerFunc {
	th := http.HandlerFunc(server.HTTPHandler)
	return func(c echo.Context) error {
		th.ServeHTTP(c.Response().Writer, c.Request())
		return nil
	}
}

func exampleMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Could capture active users here
		fmt.Println("Hello wrapped")
		next.ServeHTTP(w, r)
	})
}

func echoExampleMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("Hello echo")
		user := c.Get("loggedInUser").(uuid.User)
		fmt.Println(user.Uuid)
		return next(c)
	}
}
