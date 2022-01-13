package policy_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestReferencePolicy(t *testing.T) {
	test.RunSpecs(t, "Gateway API ReferencePolicy support")
}
