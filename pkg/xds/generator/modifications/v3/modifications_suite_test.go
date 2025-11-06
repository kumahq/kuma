package v3_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
	_ "github.com/kumahq/kuma/v2/pkg/xds/envoy"
)

func TestModifications(t *testing.T) {
	test.RunSpecs(t, "Modifications V3 Suite")
}
