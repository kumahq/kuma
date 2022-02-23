package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/membership/kubernetes"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Membership Kubernetes Suite")
}

var _ = Describe("Test Kubernetes", kubernetes.MembershipKubernetes)
