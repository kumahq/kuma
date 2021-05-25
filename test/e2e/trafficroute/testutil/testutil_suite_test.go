package testutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestApproximatelyMatcher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApproximatelyEqual Matcher Suite")
}
