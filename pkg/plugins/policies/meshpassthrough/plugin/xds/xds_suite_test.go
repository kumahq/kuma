package xds_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestPlugin(t *testing.T) {
	test.RunSpecs(t, "MeshPassthrough XDS")
}
