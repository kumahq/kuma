package hostname_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestHostnameGenerator(t *testing.T) {
	test.RunSpecs(t, "MeshService HostnameGenerator Suite")
}
