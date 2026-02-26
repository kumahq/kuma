package xds_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestConfigurer(t *testing.T) {
	test.RunSpecs(t, "MeshAccessLog XDS Configurer")
}
