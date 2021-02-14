package info

import (
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

type UserInfoService struct {
	userRepo   user.ManagementStorage
	logContext logrus.FieldLogger
}
