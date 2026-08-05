package xds_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestXDS(t *testing.T) {
	test.RunSpecs(t, "MeshAccessLog xDS Suite")
}
