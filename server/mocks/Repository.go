// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	spaced_repetition "github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// DeleteEntry provides a mock function with given fields: userUUID, UUID
func (_m *Repository) DeleteEntry(userUUID string, UUID string) error {
	ret := _m.Called(userUUID, UUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(userUUID, UUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetEntries provides a mock function with given fields: userUUID
func (_m *Repository) GetEntries(userUUID string) ([]interface{}, error) {
	ret := _m.Called(userUUID)

	var r0 []interface{}
	if rf, ok := ret.Get(0).(func(string) []interface{}); ok {
		r0 = rf(userUUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetEntry provides a mock function with given fields: userUUID, UUID
func (_m *Repository) GetEntry(userUUID string, UUID string) (interface{}, error) {
	ret := _m.Called(userUUID, UUID)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string, string) interface{}); ok {
		r0 = rf(userUUID, UUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(userUUID, UUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNext provides a mock function with given fields: userUUID
func (_m *Repository) GetNext(userUUID string) (interface{}, error) {
	ret := _m.Called(userUUID)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string) interface{}); ok {
		r0 = rf(userUUID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(userUUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveEntry provides a mock function with given fields: entry
func (_m *Repository) SaveEntry(entry spaced_repetition.SpacedRepetitionEntry) error {
	ret := _m.Called(entry)

	var r0 error
	if rf, ok := ret.Get(0).(func(spaced_repetition.SpacedRepetitionEntry) error); ok {
		r0 = rf(entry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateEntry provides a mock function with given fields: entry
func (_m *Repository) UpdateEntry(entry spaced_repetition.SpacedRepetitionEntry) error {
	ret := _m.Called(entry)

	var r0 error
	if rf, ok := ret.Get(0).(func(spaced_repetition.SpacedRepetitionEntry) error); ok {
		r0 = rf(entry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
