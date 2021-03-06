package api

import (
	"github.com/freshteapot/learnalist-api/server/api/models"
	"github.com/freshteapot/learnalist-api/server/pkg/acl"
	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/freshteapot/learnalist-api/server/pkg/user"
	"github.com/sirupsen/logrus"
)

// m exposing the data abstraction layer

type Manager struct {
	Datastore      models.Datastore
	UserManagement user.Management
	Acl            acl.Acl
	DatabaseName   string
	logger         logrus.FieldLogger
	insights       event.Insights
}

func NewManager(
	datastore models.Datastore,
	userManagement user.Management,
	acl acl.Acl,
	databaseName string,

	logger logrus.FieldLogger,
) *Manager {
	return &Manager{
		Datastore:      datastore,
		UserManagement: userManagement,
		Acl:            acl,
		DatabaseName:   databaseName,
		logger:         logger,
		insights:       event.NewInsights(logger),
	}
}

// HACK remove once I have tamed this madness
func (m *Manager) SetLogger(logger logrus.FieldLogger) {
	m.logger = logger
}
