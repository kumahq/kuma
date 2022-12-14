package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficpermission/kubernetes"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Traffic Permission Kubernetes Suite")
}

var _ = Describe("Traffic Permission on Kubernetes", Label("job-2"), kubernetes.TrafficPermission)
