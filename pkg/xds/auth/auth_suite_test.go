package auth_test

import (
	"testing"

	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestAuth(t *testing.T) {
	test.RunSpecs(t, "XDS Auth Suite")
}
