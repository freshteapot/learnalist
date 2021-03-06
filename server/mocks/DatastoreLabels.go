// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	label "github.com/freshteapot/learnalist-api/server/api/label"
	mock "github.com/stretchr/testify/mock"
)

// DatastoreLabels is an autogenerated mock type for the DatastoreLabels type
type DatastoreLabels struct {
	mock.Mock
}

// Labels provides a mock function with given fields:
func (_m *DatastoreLabels) Labels() label.LabelReadWriter {
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

// RemoveUserLabel provides a mock function with given fields: _a0, uuid
func (_m *DatastoreLabels) RemoveUserLabel(_a0 string, uuid string) error {
	ret := _m.Called(_a0, uuid)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(_a0, uuid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
