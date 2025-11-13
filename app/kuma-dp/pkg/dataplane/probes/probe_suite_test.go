package probes_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestVirtualProbes(t *testing.T) {
	test.RunSpecs(t, "Application probe proxy suite")
}
