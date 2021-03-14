package api

import (
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/oauth"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	userRegisterKey             string
	aclRepo                     acl.Acl
	userManagement              user.Management
	userWithUsernameAndPassword user.UserWithUsernameAndPassword
	userInfoRepo                user.UserInfoRepository
	userFromIDP                 user.UserFromIDP
	userSession                 user.Session
	oauthHandlers               oauth.Handlers
	logContext                  logrus.FieldLogger
}

// @openapi.path.tag: user
func NewService(
	oauthHandlers oauth.Handlers,
	aclRepo acl.Acl,
	userManagement user.Management,
	userFromIDP user.UserFromIDP,
	userSession user.Session,
	userWithUsernameAndPassword user.UserWithUsernameAndPassword,
	userInfoRepo user.UserInfoRepository,
	userRegisterKey string,
	logContext logrus.FieldLogger,
) UserService {

	s := UserService{
		oauthHandlers:               oauthHandlers,
		aclRepo:                     aclRepo,
		userManagement:              userManagement,
		userFromIDP:                 userFromIDP,
		userSession:                 userSession,
		userWithUsernameAndPassword: userWithUsernameAndPassword,
		userInfoRepo:                userInfoRepo,
		userRegisterKey:             userRegisterKey,
		logContext:                  logContext,
	}

	event.GetBus().Subscribe(event.TopicMonolog, "userService", s.OnEvent)
	return s
}
