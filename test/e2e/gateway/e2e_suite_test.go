package gateway_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/gateway"
)

var _ = Describe("Test Gateway on Universal", gateway.GatewayOnUniversal)
var _ = Describe("Test Gateway on Kubernetes", gateway.GatewayOnKubernetes)
var _ = Describe("Test Gateway on Kubernetes with HELM", gateway.GatewayHELM)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Gateway Suite")
}
