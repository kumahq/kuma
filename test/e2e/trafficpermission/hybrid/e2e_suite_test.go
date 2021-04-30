package hybrid_test

import (
	"testing"

	"github.com/kumahq/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2ETrafficPermissionHybrid(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "Traffic Permission Hybrid Suite")
	} else {
		t.SkipNow()
	}
}
