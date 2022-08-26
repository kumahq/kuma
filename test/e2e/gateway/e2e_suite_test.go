package gateway_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Gateway Suite")
}
