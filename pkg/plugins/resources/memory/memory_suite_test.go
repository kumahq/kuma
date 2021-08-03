package memory_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestClient(t *testing.T) {
	test.RunSpecs(t, "In-memory ResourceStore Suite")
}
