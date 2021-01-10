// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	stan "github.com/nats-io/stan.go"
	mock "github.com/stretchr/testify/mock"
)

// NatsSubscriber is an autogenerated mock type for the NatsSubscriber type
type NatsSubscriber struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *NatsSubscriber) Close() {
	_m.Called()
}

// Subscribe provides a mock function with given fields: topic, sc
func (_m *NatsSubscriber) Subscribe(topic string, sc stan.Conn) error {
	ret := _m.Called(topic, sc)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, stan.Conn) error); ok {
		r0 = rf(topic, sc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
