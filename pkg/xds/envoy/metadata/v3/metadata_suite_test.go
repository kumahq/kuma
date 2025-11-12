package envoy_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestMetadata(t *testing.T) {
	test.RunSpecs(t, "Envoy Metadata V3 Suite")
}
