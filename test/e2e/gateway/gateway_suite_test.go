package gateway

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestE2EGateway(t *testing.T) {
	test.RunSpecs(t, "E2E Gateway Suite")
}
