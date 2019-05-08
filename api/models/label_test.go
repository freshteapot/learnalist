package models

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	resetDatabase()
}

func TestPostUserLabel(t *testing.T) {
	resetDatabase()
	a := NewUserLabel("label1", "user2")
	statusCode, _ := dal.PostUserLabel(a)
	assert.Equal(t, http.StatusCreated, statusCode)

	statusCode, _ = dal.PostUserLabel(a)
	assert.Equal(t, http.StatusOK, statusCode)

	b := NewUserLabel("label_123456789_123456789_car_boat", "user2")
	statusCode, err := dal.PostUserLabel(b)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, ValidationWarningLabelToLong, err.Error())
}

func TestPostAlistLabel(t *testing.T) {
	resetDatabase()
	a := NewAlistLabel("label1", "u:123", "u:456")
	statusCode, _ := dal.PostAlistLabel(a)
	assert.Equal(t, http.StatusCreated, statusCode)

	statusCode, _ = dal.PostAlistLabel(a)
	assert.Equal(t, http.StatusOK, statusCode)

	b := NewAlistLabel("label_123456789_123456789_car_boat", "u:123", "u:456")
	statusCode, err := dal.PostAlistLabel(b)
	assert.Equal(t, http.StatusBadRequest, statusCode)
	assert.Equal(t, ValidationWarningLabelToLong, err.Error())
}

func TestGetUserLabels(t *testing.T) {
	resetDatabase()
	// Testing for an empty response
	emptyList := make([]string, 0)
	labels, _ := dal.GetUserLabels("u:123")
	assert.Equal(t, emptyList, labels)

	a := NewUserLabel("label1", "u:123")
	statusCode, _ := dal.PostUserLabel(a)
	assert.Equal(t, http.StatusCreated, statusCode)

	labels, _ = dal.GetUserLabels("u:123")

	b := NewAlistLabel("label2", "u:123", "u:456")
	statusCode, _ = dal.PostAlistLabel(b)
	assert.Equal(t, http.StatusCreated, statusCode)
	// We add the same one, just to make sure we only get 2 back.
	statusCode, _ = dal.PostAlistLabel(b)
	labels, _ = dal.GetUserLabels("u:123")
	assert.Equal(t, 2, len(labels))

	// Test remove
	dal.RemoveUserLabel("label2", "u:123")
	labels, _ = dal.GetUserLabels("u:123")
	assert.Equal(t, 1, len(labels))
	dal.RemoveUserLabel("label1", "u:123")
	labels, _ = dal.GetUserLabels("u:123")
	assert.Equal(t, 0, len(labels))
}
