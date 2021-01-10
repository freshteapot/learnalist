// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// AclWriterChallenge is an autogenerated mock type for the AclWriterChallenge type
type AclWriterChallenge struct {
	mock.Mock
}

// DeleteChallenge provides a mock function with given fields: extUUID
func (_m *AclWriterChallenge) DeleteChallenge(extUUID string) error {
	ret := _m.Called(extUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(extUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GrantUserChallengeWriteAccess provides a mock function with given fields: extUUID, userUUID
func (_m *AclWriterChallenge) GrantUserChallengeWriteAccess(extUUID string, userUUID string) error {
	ret := _m.Called(extUUID, userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(extUUID, userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// MakeChallengePrivate provides a mock function with given fields: extUUID, userUUID
func (_m *AclWriterChallenge) MakeChallengePrivate(extUUID string, userUUID string) error {
	ret := _m.Called(extUUID, userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(extUUID, userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeUserChallengeWriteAccess provides a mock function with given fields: extUUID, userUUID
func (_m *AclWriterChallenge) RevokeUserChallengeWriteAccess(extUUID string, userUUID string) error {
	ret := _m.Called(extUUID, userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(extUUID, userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ShareChallengeWithPublic provides a mock function with given fields: extUUID
func (_m *AclWriterChallenge) ShareChallengeWithPublic(extUUID string) error {
	ret := _m.Called(extUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(extUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
