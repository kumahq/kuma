package kubernetes_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EKICKubernetes(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "KIC Kubernetes Suite")
	} else {
		t.SkipNow()
	}
}
