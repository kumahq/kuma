package v3_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
	_ "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

func TestModifications(t *testing.T) {
	test.RunSpecs(t, "Modifications V3 Suite")
}
