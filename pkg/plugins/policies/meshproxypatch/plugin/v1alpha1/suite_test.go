package v1alpha1_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
	_ "github.com/kumahq/kuma/v2/pkg/xds/envoy"
)

func TestModifications(t *testing.T) {
	test.RunSpecs(t, "MeshProxyPatch Plugin Suite")
}
