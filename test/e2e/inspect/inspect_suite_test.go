package inspect_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/inspect"
)

var _ = Describe("Test Inspect API on Universal", inspect.Universal)

// Disabling tests to implement them later using `kumactl` instead of `wget`
var _ = XDescribe("Test Inspect API on Kubernetes Standalone", inspect.KubernetesStandalone)
var _ = XDescribe("Test Inspect API on Kubernetes Multizone", inspect.KubernetesMultizone)

func TestE2EInspectAPI(t *testing.T) {
	test.RunSpecs(t, "E2E Inspect API Suite")
}
