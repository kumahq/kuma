package virtualoutbound_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/test/e2e/virtualoutbound"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E VirtualOutbound Suite")
}

var _ = Describe("Test VirtualOutbound on K8s", Label("job-0"), virtualoutbound.VirtualOutboundOnK8s)
