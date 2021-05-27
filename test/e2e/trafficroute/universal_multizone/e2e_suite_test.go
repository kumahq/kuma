package universal_multizone_test

import (
	"testing"

	"github.com/kumahq/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2ETrafficRouteUniversalMultizone(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "Traffic Route Universal Multizone Suite")
	} else {
		t.SkipNow()
	}
}
