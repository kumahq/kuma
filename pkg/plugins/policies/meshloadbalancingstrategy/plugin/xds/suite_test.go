package xds_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestMeshLoadBalancingStrategyXDS(t *testing.T) {
	test.RunSpecs(t, "MeshLoadBalancingStrategy XDS Suite")
}
