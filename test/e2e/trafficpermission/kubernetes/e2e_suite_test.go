package kubernetes_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficpermission/kubernetes"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ETrafficPermissionKubernetes(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Traffic Permission Kubernetes Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Traffic Permission on Kubernetes", kubernetes.TrafficPermission)
