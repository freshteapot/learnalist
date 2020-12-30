package app_settings_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuthenticate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "App Settings Test Suite")
}
