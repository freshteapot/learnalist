// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	assets "github.com/freshteapot/learnalist-api/server/pkg/assets"
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

// GetEntry provides a mock function with given fields: UUID
func (_m *Repository) GetEntry(UUID string) (assets.AssetEntry, error) {
	ret := _m.Called(UUID)

	var r0 assets.AssetEntry
	if rf, ok := ret.Get(0).(func(string) assets.AssetEntry); ok {
		r0 = rf(UUID)
	} else {
		r0 = ret.Get(0).(assets.AssetEntry)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(UUID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SaveEntry provides a mock function with given fields: entry
func (_m *Repository) SaveEntry(entry assets.AssetEntry) error {
	ret := _m.Called(entry)

	var r0 error
	if rf, ok := ret.Get(0).(func(assets.AssetEntry) error); ok {
		r0 = rf(entry)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
