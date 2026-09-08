package generator_test

import (
	"testing"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestGenerator(t *testing.T) {
	test.RunSpecs(t, "Generator Suite")
}

// DummyInternalAddresses are used when the internal addresses should not be used when generating Envoy config
var DummyInternalAddresses = []core_xds.InternalAddress{
	{AddressPrefix: "100.64.0.0", PrefixLen: 16},
}
