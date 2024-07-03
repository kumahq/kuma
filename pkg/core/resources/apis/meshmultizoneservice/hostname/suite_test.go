package hostname_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestHostnameGenerator(t *testing.T) {
	test.RunSpecs(t, "MeshMultiZoneServiceHostnameGenerator Suite")
}
