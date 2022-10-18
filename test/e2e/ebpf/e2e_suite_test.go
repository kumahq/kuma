package ebpf_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/ebpf"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Ebpf Suite")
}

var _ = Describe("Test Cleanup eBPF", Label("job-0"), Label("arm-not-supported"), Label("old-k3d-not-supported"), ebpf.CleanupEbpfConfigFromNode, Ordered)
