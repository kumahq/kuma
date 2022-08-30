package virtualoutbound_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/virtualoutbound"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E VirtualOutbound Suite")
}

var _ = Describe("Test VirtualOutbound on K8s", Label("job-1"), virtualoutbound.VirtualOutboundOnK8s)
