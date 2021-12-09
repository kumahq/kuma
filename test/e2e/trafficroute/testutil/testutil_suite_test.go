package testutil_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestApproximatelyMatcher(t *testing.T) {
	test.RunSpecs(t, "ApproximatelyEqual Matcher Suite")
}
