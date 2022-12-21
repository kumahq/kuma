package xds_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestPlugin(t *testing.T) {
	test.RunSpecs(t, "MeshHealthCheck configurer")
}
