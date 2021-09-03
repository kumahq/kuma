package virtualoutbound_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/virtualoutbound"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ERetry(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E VirtualOutbound Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test VirtualOutbound on Universal", virtualoutbound.VirtualOutboundOnUniversal)
var _ = Describe("Test VirtualOutbound on K8s", virtualoutbound.VirtualOutboundOnK8s)
