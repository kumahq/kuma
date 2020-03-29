package faultinjections

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestMatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "FaultInjection Matcher Suite")
}
