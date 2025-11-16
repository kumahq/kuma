package faultinjections

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestMatcher(t *testing.T) {
	test.RunSpecs(t, "FaultInjection Matcher Suite")
}
