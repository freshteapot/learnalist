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
	newUserA, err = dal.InsertNewUser(loginUser)
	assert.NoError(t, err)

	newUserB, err = dal.GetUserByCredentials(loginUser)
	assert.Equal(t, newUserA, newUserB)

	_, err = dal.InsertNewUser(loginUser)
	assert.Equal(t, i18n.UserInsertUsernameExists, err.Error())
	/*
		loginUser.Password = "change"
		newUserC, err := dal.InsertNewUser(loginUser)
		fmt.Println(err)
		fmt.Println(newUserC)
		newUserC, err = dal.InsertNewUser(loginUser)
		fmt.Println(err)
		fmt.Println(newUserC)
		assert.Equal(t, i18n.UserInsertAlreadyExistsPasswordNotMatch, err.Error())

		loginUser.Username = "iamanotheruser"
		_, err = dal.GetUserByCredentials(loginUser)
		assert.Equal(t, i18n.DatabaseLookupNotFound, err.Error())
	*/
	loginUser.Username = "iamanotheruser"
	_, err = dal.GetUserByCredentials(loginUser)
	assert.Equal(t, i18n.DatabaseLookupNotFound, err.Error())
}
