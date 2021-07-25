package v3_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
)

func TestModifications(t *testing.T) {
	test.RunSpecs(t, "Modifications V3 Suite")
}
