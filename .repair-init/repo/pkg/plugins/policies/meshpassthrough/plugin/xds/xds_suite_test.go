package xds_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestPlugin(t *testing.T) {
	test.RunSpecs(t, "MeshPassthrough XDS")
}
