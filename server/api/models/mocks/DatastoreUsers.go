// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// DatastoreUsers is an autogenerated mock type for the DatastoreUsers type
type DatastoreUsers struct {
	mock.Mock
}

// UserExists provides a mock function with given fields: userUUID
func (_m *DatastoreUsers) UserExists(userUUID string) bool {
	ret := _m.Called(userUUID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(userUUID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}
