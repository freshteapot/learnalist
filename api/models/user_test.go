package models

import (
	"testing"

	"github.com/freshteapot/learnalist-api/api/authenticate"
	"github.com/freshteapot/learnalist-api/api/i18n"
	"github.com/freshteapot/learnalist-api/api/uuid"
	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestUser(t *testing.T) {
	resetDatabase()
	var newUserA *uuid.User
	var newUserB *uuid.User
	var err error

	loginUser := authenticate.LoginUser{
		Username: "iamuser",
		Password: "iampassword",
	}
	// Insert new user
	newUserA, err = dal.InsertNewUser(loginUser)
	assert.NoError(t, err)
	// Confirm the user uuid is the same.
	newUserB, err = dal.GetUserByCredentials(loginUser)
	assert.Equal(t, newUserA, newUserB)
	// Insert the same user and confirm it is rejected.
	_, err = dal.InsertNewUser(loginUser)
	assert.Equal(t, i18n.UserInsertUsernameExists, err.Error())
	// Confirm getting a user that is not the system is handled.
	loginUser.Username = "iamanotheruser"
	_, err = dal.GetUserByCredentials(loginUser)
	assert.Equal(t, i18n.DatabaseLookupNotFound, err.Error())
}
