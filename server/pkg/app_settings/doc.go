package app_settings

import (
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

type AppSettingsService struct {
	userInfoRepo user.UserInfoRepository
	logContext   logrus.FieldLogger
}
