package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/healthcheck/kubernetes"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EHealthCheckKubernetes(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Health Check Kubernetes Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Virtual Probes on Kubernetes", kubernetes.VirtualProbes)
