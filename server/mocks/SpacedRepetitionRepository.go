// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	spaced_repetition "github.com/freshteapot/learnalist-api/server/pkg/spaced_repetition"
	mock "github.com/stretchr/testify/mock"
)

// SpacedRepetitionRepository is an autogenerated mock type for the SpacedRepetitionRepository type
type SpacedRepetitionRepository struct {
	mock.Mock
}

// DeleteByUser provides a mock function with given fields: userUUID
func (_m *SpacedRepetitionRepository) DeleteByUser(userUUID string) error {
	ret := _m.Called(userUUID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(userUUID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteEntry provides a mock function with given fields: userUUID, UUID
func (_m *SpacedRepetitionRepository) DeleteEntry(userUUID string, UUID string) error {
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
func (_m *SpacedRepetitionRepository) GetEntries(userUUID string) ([]interface{}, error) {
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
func (_m *SpacedRepetitionRepository) GetEntry(userUUID string, UUID string) (interface{}, error) {
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
func (_m *SpacedRepetitionRepository) GetNext(userUUID string) (spaced_repetition.SpacedRepetitionEntry, error) {
	ret := _m.Called(userUUID)

	var r0 spaced_repetition.SpacedRepetitionEntry
	if rf, ok := ret.Get(0).(func(string) spaced_repetition.SpacedRepetitionEntry); ok {
		r0 = rf(userUUID)
	} else {
		r0 = ret.Get(0).(spaced_repetition.SpacedRepetitionEntry)
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
func (_m *SpacedRepetitionRepository) SaveEntry(entry spaced_repetition.SpacedRepetitionEntry) error {
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
func (_m *SpacedRepetitionRepository) UpdateEntry(entry spaced_repetition.SpacedRepetitionEntry) error {
	ret := _m.Called(entry)

	var r0 error
	if rf, ok := ret.Get(0).(func(spaced_repetition.SpacedRepetitionEntry) error); ok {
		r0 = rf(entry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
