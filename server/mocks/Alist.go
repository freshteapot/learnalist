// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	alist "github.com/freshteapot/learnalist-api/server/api/alist"
	mock "github.com/stretchr/testify/mock"
)

// Alist is an autogenerated mock type for the Alist type
type Alist struct {
	mock.Mock
}

// GetPublicLists provides a mock function with given fields:
func (_m *Alist) GetPublicLists() []alist.ShortInfo {
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

// Insert provides a mock function with given fields: aList
func (_m *Alist) Insert(aList alist.Alist) error {
	ret := _m.Called(aList)

	var r0 error
	if rf, ok := ret.Get(0).(func(alist.Alist) error); ok {
		r0 = rf(aList)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Remove provides a mock function with given fields: alistUUID, userUUID
func (_m *Alist) Remove(alistUUID string, userUUID string) error {
	ret := _m.Called(alistUUID, userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(alistUUID, userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: aList
func (_m *Alist) Update(aList alist.Alist) error {
	ret := _m.Called(aList)

	var r0 error
	if rf, ok := ret.Get(0).(func(alist.Alist) error); ok {
		r0 = rf(aList)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
