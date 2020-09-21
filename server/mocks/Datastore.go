// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	alist "github.com/freshteapot/learnalist-api/server/api/alist"
	label "github.com/freshteapot/learnalist-api/server/api/label"

	mock "github.com/stretchr/testify/mock"

	oauth "github.com/freshteapot/learnalist-api/server/pkg/oauth"

	user "github.com/freshteapot/learnalist-api/server/pkg/user"
)

// Datastore is an autogenerated mock type for the Datastore type
type Datastore struct {
	mock.Mock
}

// GetAlist provides a mock function with given fields: uuid
func (_m *Datastore) GetAlist(uuid string) (alist.Alist, error) {
	ret := _m.Called(uuid)

	var r0 alist.Alist
	if rf, ok := ret.Get(0).(func(string) alist.Alist); ok {
		r0 = rf(uuid)
	} else {
		r0 = ret.Get(0).(alist.Alist)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uuid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAllListsByUser provides a mock function with given fields: userUUID
func (_m *Datastore) GetAllListsByUser(userUUID string) []alist.ShortInfo {
	ret := _m.Called(userUUID)

	var r0 []alist.ShortInfo
	if rf, ok := ret.Get(0).(func(string) []alist.ShortInfo); ok {
		r0 = rf(userUUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]alist.ShortInfo)
		}
	}

	return r0
}

// GetListsByUserWithFilters provides a mock function with given fields: uuid, labels, listType
func (_m *Datastore) GetListsByUserWithFilters(uuid string, labels string, listType string) []alist.Alist {
	ret := _m.Called(uuid, labels, listType)

	var r0 []alist.Alist
	if rf, ok := ret.Get(0).(func(string, string, string) []alist.Alist); ok {
		r0 = rf(uuid, labels, listType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]alist.Alist)
		}
	}

	return r0
}

// GetPublicLists provides a mock function with given fields:
func (_m *Datastore) GetPublicLists() []alist.ShortInfo {
	ret := _m.Called()

	var r0 []alist.ShortInfo
	if rf, ok := ret.Get(0).(func() []alist.ShortInfo); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]alist.ShortInfo)
		}
	}

	return r0
}

// Labels provides a mock function with given fields:
func (_m *Datastore) Labels() label.LabelReadWriter {
	ret := _m.Called()

	var r0 label.LabelReadWriter
	if rf, ok := ret.Get(0).(func() label.LabelReadWriter); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(label.LabelReadWriter)
		}
	}

	return r0
}

// OAuthHandler provides a mock function with given fields:
func (_m *Datastore) OAuthHandler() oauth.OAuthReadWriter {
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

// RemoveAlist provides a mock function with given fields: alistUUID, userUUID
func (_m *Datastore) RemoveAlist(alistUUID string, userUUID string) error {
	ret := _m.Called(alistUUID, userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(alistUUID, userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RemoveUserLabel provides a mock function with given fields: _a0, uuid
func (_m *Datastore) RemoveUserLabel(_a0 string, uuid string) error {
	ret := _m.Called(_a0, uuid)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(_a0, uuid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SaveAlist provides a mock function with given fields: method, aList
func (_m *Datastore) SaveAlist(method string, aList alist.Alist) (alist.Alist, error) {
	ret := _m.Called(method, aList)

	var r0 alist.Alist
	if rf, ok := ret.Get(0).(func(string, alist.Alist) alist.Alist); ok {
		r0 = rf(method, aList)
	} else {
		r0 = ret.Get(0).(alist.Alist)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, alist.Alist) error); ok {
		r1 = rf(method, aList)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UserExists provides a mock function with given fields: userUUID
func (_m *Datastore) UserExists(userUUID string) bool {
	ret := _m.Called(userUUID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(userUUID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// UserFromIDP provides a mock function with given fields:
func (_m *Datastore) UserFromIDP() user.UserFromIDP {
	ret := _m.Called()

	var r0 user.UserFromIDP
	if rf, ok := ret.Get(0).(func() user.UserFromIDP); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(user.UserFromIDP)
		}
	}

	return r0
}

// UserSession provides a mock function with given fields:
func (_m *Datastore) UserSession() user.Session {
	ret := _m.Called()

	var r0 user.Session
	if rf, ok := ret.Get(0).(func() user.Session); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(user.Session)
		}
	}

	return r0
}

// UserWithUsernameAndPassword provides a mock function with given fields:
func (_m *Datastore) UserWithUsernameAndPassword() user.UserWithUsernameAndPassword {
	ret := _m.Called()

	var r0 user.UserWithUsernameAndPassword
	if rf, ok := ret.Get(0).(func() user.UserWithUsernameAndPassword); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(user.UserWithUsernameAndPassword)
		}
	}

	return r0
}
