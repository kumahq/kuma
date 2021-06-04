package universal_standalone_test

import (
	"testing"

	"github.com/kumahq/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2ETrafficRouteUniversalStandalone(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "Traffic Route Universal Standalone Suite")
	} else {
		t.SkipNow()
	}
}
