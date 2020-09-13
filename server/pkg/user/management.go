package user

import (
	"fmt"

	"github.com/freshteapot/learnalist-api/server/pkg/event"
	"github.com/sirupsen/logrus"
)

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
	found, _ := m.storage.FindUserUUID(userUUID)
	if len(found) == 0 {
		return ErrNotFound
	}

	// This code is not deleting from the database
	lists, err := m.storage.GetLists(userUUID)
	if err != nil {
		return err
	}

	// Remove from the site
	for _, listUUID := range lists {
		// fmt.Printf("Remove list %s from static site\n", listUUID)
		m.site.DeleteList(listUUID)
		// fmt.Printf("Remove list %s from db\n", listUUID)
		err = m.storage.DeleteList(listUUID)
		if err != nil {
			fmt.Println("DeleteUser", err)
			return err
		}
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
