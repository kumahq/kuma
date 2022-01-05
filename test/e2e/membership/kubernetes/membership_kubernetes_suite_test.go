package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/membership/kubernetes"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EMembershipKubernetes(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Membership Kubernetes Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Kubernetes", kubernetes.MembershipKubernetes)
