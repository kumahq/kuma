package universal_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestUniversal(t *testing.T) {
	test.RunSpecs(t, "XDS Auth Universal Suite")
}
