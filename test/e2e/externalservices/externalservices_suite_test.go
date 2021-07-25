package externalservices_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EExternalServices(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E External Services Suite")
	} else {
		t.SkipNow()
	}
}
