package app_settings

import (
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

type AppSettingsService struct {
	userRepo   user.ManagementStorage
	logContext logrus.FieldLogger
}
