// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// AclReaderList is an autogenerated mock type for the AclReaderList type
type AclReaderList struct {
	mock.Mock
}

// HasUserListReadAccess provides a mock function with given fields: alistUUID, userUUID
func (_m *AclReaderList) HasUserListReadAccess(alistUUID string, userUUID string) (bool, error) {
	ret := _m.Called(alistUUID, userUUID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(alistUUID, userUUID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(alistUUID, userUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasUserListWriteAccess provides a mock function with given fields: alistUUID, userUUID
func (_m *AclReaderList) HasUserListWriteAccess(alistUUID string, userUUID string) (bool, error) {
	ret := _m.Called(alistUUID, userUUID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string) bool); ok {
		r0 = rf(alistUUID, userUUID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(alistUUID, userUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsListAvailableToFriends provides a mock function with given fields: alistUUID
func (_m *AclReaderList) IsListAvailableToFriends(alistUUID string) (bool, error) {
	ret := _m.Called(alistUUID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(alistUUID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(alistUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsListPrivate provides a mock function with given fields: alistUUID
func (_m *AclReaderList) IsListPrivate(alistUUID string) (bool, error) {
	ret := _m.Called(alistUUID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(alistUUID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(alistUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsListPublic provides a mock function with given fields: alistUUID
func (_m *AclReaderList) IsListPublic(alistUUID string) (bool, error) {
	ret := _m.Called(alistUUID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(alistUUID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(alistUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListIsSharedWith provides a mock function with given fields: alistUUID
func (_m *AclReaderList) ListIsSharedWith(alistUUID string) (string, error) {
	ret := _m.Called(alistUUID)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(alistUUID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(alistUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
