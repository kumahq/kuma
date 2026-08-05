package memory_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestClient(t *testing.T) {
	test.RunSpecs(t, "In-memory ResourceStore Suite")
}
