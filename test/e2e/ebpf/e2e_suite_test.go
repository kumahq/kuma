package ebpf_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/ebpf"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Ebpf Suite")
}

// Tests fail on github CI with:
//
//	Type: "Warning",
//	Object: "Pod/test-server-599d497f-t4g5q",
//	Reason: "Failed",
//	Message: "Error: failed to generate container \"596f3396a0f42c6a554a95d6b1599d7ed3682ba2f95ae317b307c493891dd084\" spec: failed to generate spec: path \"/sys/fs/bpf\" is mounted on \"/sys\" but it is not a shared mount",
var _ = PDescribe("Test Cleanup eBPF", Label("job-0"), Label("arm-not-supported"), Label("legacy-k3s-not-supported"), ebpf.CleanupEbpfConfigFromNode, Ordered)
