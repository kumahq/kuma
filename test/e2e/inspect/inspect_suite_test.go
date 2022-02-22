package inspect_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/inspect"
)

var _ = Describe("Test Inspect API on Universal", inspect.Universal)

var _ = Describe("Test Inspect API on Kubernetes Standalone", inspect.KubernetesStandalone)
var _ = Describe("Test Inspect API on Kubernetes Multizone", inspect.KubernetesMultizone)

func TestE2EInspectAPI(t *testing.T) {
	test.RunSpecs(t, "E2E Inspect API Suite")
}
