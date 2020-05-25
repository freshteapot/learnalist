package user

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

type ManagementStorage interface {
	FindUserUUID(search string) ([]string, error)
	GetLists(userUUID string) ([]string, error)
	DeleteUser(userUUID string) error
}

type ManagementSite interface {
	DeleteList(listUUID string) error
	DeleteUser(userUUID string) error
}

type Management interface {
	FindUser(search string) ([]string, error)
	DeleteUser(userUUID string) error
}

type management struct {
	storage  ManagementStorage
	site     ManagementSite
	insights event.Insights
}

func NewManagement(storage ManagementStorage, site ManagementSite, insights event.Insights) management {
	return management{
		storage:  storage,
		site:     site,
		insights: insights,
	}
}

// FindUser Find the user uuid based on the search string
func (m management) FindUser(search string) ([]string, error) {
	return m.storage.FindUserUUID(search)
}

func (m management) DeleteUser(userUUID string) error {
	lists, err := m.storage.GetLists(userUUID)
	if err != nil {
		return err
	}

	// Remove from the site
	for _, listUUID := range lists {
		fmt.Printf("Remove list %s from static site", listUUID)
		m.site.DeleteList(listUUID)
	}

	err = m.site.DeleteUser(userUUID)
	if err != nil {
		return err
	}

	err = m.storage.DeleteUser(userUUID)
	if err != nil {
		return err
	}

	// TODO event that this happened
	m.insights.Event(logrus.Fields{
		"event":     event.UserDeleted,
		"user_uuid": userUUID,
	})
	return nil
}
