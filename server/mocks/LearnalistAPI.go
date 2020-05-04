// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	alist "github.com/freshteapot/learnalist-api/server/api/alist"
	api "github.com/freshteapot/learnalist-api/server/api/api"

	io "io"

	mock "github.com/stretchr/testify/mock"
)

// LearnalistAPI is an autogenerated mock type for the LearnalistAPI type
type LearnalistAPI struct {
	mock.Mock
}

// DeleteAlist provides a mock function with given fields: uuid
func (_m *LearnalistAPI) DeleteAlist(uuid string) (int, error) {
	ret := _m.Called(uuid)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(uuid)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uuid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAlist provides a mock function with given fields: uuid
func (_m *LearnalistAPI) GetAlist(uuid string) (int, alist.Alist, error) {
	ret := _m.Called(uuid)

	var r0 int
	if rf, ok := ret.Get(0).(func(string) int); ok {
		r0 = rf(uuid)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 alist.Alist
	if rf, ok := ret.Get(1).(func(string) alist.Alist); ok {
		r1 = rf(uuid)
	} else {
		r1 = ret.Get(1).(alist.Alist)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string) error); ok {
		r2 = rf(uuid)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// GetRoot provides a mock function with given fields:
func (_m *LearnalistAPI) GetRoot() (api.HttpResponseMessage, error) {
	ret := _m.Called()

	var r0 api.HttpResponseMessage
	if rf, ok := ret.Get(0).(func() api.HttpResponseMessage); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(api.HttpResponseMessage)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetVersion provides a mock function with given fields:
func (_m *LearnalistAPI) GetVersion() (api.HttpGetVersionResponse, error) {
	ret := _m.Called()

	var r0 api.HttpGetVersionResponse
	if rf, ok := ret.Get(0).(func() api.HttpGetVersionResponse); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(api.HttpGetVersionResponse)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PostAlist provides a mock function with given fields: body
func (_m *LearnalistAPI) PostAlist(body io.Reader) (int, alist.Alist, error) {
	ret := _m.Called(body)

	var r0 int
	if rf, ok := ret.Get(0).(func(io.Reader) int); ok {
		r0 = rf(body)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 alist.Alist
	if rf, ok := ret.Get(1).(func(io.Reader) alist.Alist); ok {
		r1 = rf(body)
	} else {
		r1 = ret.Get(1).(alist.Alist)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(io.Reader) error); ok {
		r2 = rf(body)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// PutAlist provides a mock function with given fields: uuid, body
func (_m *LearnalistAPI) PutAlist(uuid string, body io.Reader) (int, alist.Alist, error) {
	ret := _m.Called(uuid, body)

	var r0 int
	if rf, ok := ret.Get(0).(func(string, io.Reader) int); ok {
		r0 = rf(uuid, body)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 alist.Alist
	if rf, ok := ret.Get(1).(func(string, io.Reader) alist.Alist); ok {
		r1 = rf(uuid, body)
	} else {
		r1 = ret.Get(1).(alist.Alist)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(string, io.Reader) error); ok {
		r2 = rf(uuid, body)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
