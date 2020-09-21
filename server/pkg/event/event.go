package event

import (
	"github.com/sirupsen/logrus"
)

const (
	UserRegistered = "user-registered"
	UserDeleted    = "user-deleted"
)

type Insights interface {
	Event(fields logrus.Fields)
}

type insight struct {
	logger logrus.FieldLogger
}

func NewInsights(logger logrus.FieldLogger) insight {
	return insight{logger: logger}
}

func (i insight) Event(fields logrus.Fields) {
	i.logger.WithFields(fields).Info("insight")
}
