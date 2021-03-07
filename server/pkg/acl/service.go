package acl

import (
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

type AclService struct {
	repo       Acl
	logContext logrus.FieldLogger
}

func NewService(repo Acl, log logrus.FieldLogger) AclService {
	s := AclService{
		repo:       repo,
		logContext: log,
	}

	event.GetBus().Subscribe(event.TopicMonolog, "acl", s.OnEvent)
	return s
}
