package models

import (
	"testing"

	"github.com/freshteapot/learnalist-api/api/authenticate"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestGetLabel(t *testing.T) {
	var labelA *Label
	var labelB *Label
	var err error

	resetDatabase()
	loginUser := authenticate.LoginUser{
		Username: "testUser",
		Password: "password",
	}

	// Get a fake label.
	uuid := "fake"
	labelA, err = dal.GetLabel(uuid)
	assert.Error(t, err)

	// Add a label
	user, _ := dal.InsertNewUser(loginUser)

	labelA = NewLabel()
	labelA.Label = "a"
	labelA.UserUuid = user.Uuid
	dal.SaveLabel(*labelA)
	labelB, err = dal.GetLabel(labelA.Uuid)
	assert.NoError(t, err)
	assert.Equal(t, labelB, labelA)
}

func TestSaveAndGetLabel(t *testing.T) {
	resetDatabase()

	var labels []Label
	loginUser := authenticate.LoginUser{
		Username: "testUser",
		Password: "password",
	}
	user, _ := dal.InsertNewUser(loginUser)
	// Empty
	labels = dal.GetLabelsByUser(user.Uuid)
	assert.Equal(t, 0, len(labels))

	label := NewLabel()
	label.Label = "a"
	label.UserUuid = user.Uuid
	dal.SaveLabel(*label)
	labels = dal.GetLabelsByUser(user.Uuid)
	assert.Equal(t, 1, len(labels))

	// Add label with a link, making sure it doesnt add the label twice
	playList := uuid.NewPlaylist(user)
	label.AlistUuid = playList.Uuid
	dal.SaveLabel(*label)

	labels = dal.GetLabelsByUser(user.Uuid)
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, label.AlistUuid, labels[0].AlistUuid)

	// Add another label
	label = NewLabel()
	label.Label = "b"
	label.UserUuid = user.Uuid

	dal.SaveLabel(*label)
	// Confirm we have two results
	labels = dal.GetLabelsByUser(user.Uuid)
	assert.Equal(t, 2, len(labels))
}
