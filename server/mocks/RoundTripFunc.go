// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// RoundTripFunc is an autogenerated mock type for the RoundTripFunc type
type RoundTripFunc struct {
	mock.Mock
}

// Execute provides a mock function with given fields: req
func (_m *RoundTripFunc) Execute(req *http.Request) *http.Response {
	ret := _m.Called(req)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Request) *http.Response); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	return r0
}