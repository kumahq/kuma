package hostname_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestHostnameGenerator(t *testing.T) {
	test.RunSpecs(t, "MeshExternalServiceHostnameGenerator Suite")
}
