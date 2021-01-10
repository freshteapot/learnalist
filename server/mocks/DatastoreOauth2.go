// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"

	oauth "github.com/freshteapot/learnalist-api/server/pkg/oauth"
)

// DatastoreOauth2 is an autogenerated mock type for the DatastoreOauth2 type
type DatastoreOauth2 struct {
	mock.Mock
}

// OAuthHandler provides a mock function with given fields:
func (_m *DatastoreOauth2) OAuthHandler() oauth.OAuthReadWriter {
	ret := _m.Called()

	var r0 oauth.OAuthReadWriter
	if rf, ok := ret.Get(0).(func() oauth.OAuthReadWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(oauth.OAuthReadWriter)
		}
	}

	return r0
}
