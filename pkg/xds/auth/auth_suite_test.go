package auth_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestAuth(t *testing.T) {
	test.RunSpecs(t, "XDS Auth Suite")
}
