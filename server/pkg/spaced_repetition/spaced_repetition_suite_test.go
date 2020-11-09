package spaced_repetition_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuthenticate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Spaced Repetition Test Suite")
}
