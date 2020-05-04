// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	alist "github.com/freshteapot/learnalist-api/server/api/alist"

	mock "github.com/stretchr/testify/mock"
)

// HugoSiteBuilder is an autogenerated mock type for the HugoSiteBuilder type
type HugoSiteBuilder struct {
	mock.Mock
}

// Build provides a mock function with given fields:
func (_m *HugoSiteBuilder) Build() {
	_m.Called()
}

// ProcessContent provides a mock function with given fields:
func (_m *HugoSiteBuilder) ProcessContent() {
	_m.Called()
}

// Remove provides a mock function with given fields: uuid
func (_m *HugoSiteBuilder) Remove(uuid string) {
	_m.Called(uuid)
}

// WriteList provides a mock function with given fields: aList
func (_m *HugoSiteBuilder) WriteList(aList alist.Alist) {
	_m.Called(aList)
}

// WriteListsByUser provides a mock function with given fields: userUUID, lists
func (_m *HugoSiteBuilder) WriteListsByUser(userUUID string, lists []alist.ShortInfo) {
	_m.Called(userUUID, lists)
}

// WritePublicLists provides a mock function with given fields: lists
func (_m *HugoSiteBuilder) WritePublicLists(lists []alist.ShortInfo) {
	_m.Called(lists)
}