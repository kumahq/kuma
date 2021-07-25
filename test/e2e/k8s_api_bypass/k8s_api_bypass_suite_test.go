package k8s_api_bypass_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EExternalServices(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Kubernetes API Bypass")
	} else {
		t.SkipNow()
	}
}
