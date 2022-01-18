package universal_standalone_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/trafficroute/universal_standalone"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ETrafficRouteUniversalStandalone(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "Traffic Route Universal Standalone Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Test Standalone Universal deployment", universal_standalone.KumaStandalone)
