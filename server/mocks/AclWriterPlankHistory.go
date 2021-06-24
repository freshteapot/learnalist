// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// AclWriterPlankHistory is an autogenerated mock type for the AclWriterPlankHistory type
type AclWriterPlankHistory struct {
	mock.Mock
}

// MakePlankHistoryPrivate provides a mock function with given fields: userUUID
func (_m *AclWriterPlankHistory) MakePlankHistoryPrivate(userUUID string) error {
	ret := _m.Called(userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SharePlankHistoryWithPublic provides a mock function with given fields: extUUID
func (_m *AclWriterPlankHistory) SharePlankHistoryWithPublic(extUUID string) error {
	ret := _m.Called(extUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(extUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
