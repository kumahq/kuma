package v1alpha1_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	_ "github.com/kumahq/kuma/pkg/xds/envoy"
)

func TestModifications(t *testing.T) {
	test.RunSpecs(t, "MeshProxyPatch Plugin Suite")
}
