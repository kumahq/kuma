package hybrid_test

import (
	"testing"

	"github.com/kumahq/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2EHealthCheckHybrid(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "Health Check Hybrid Suite")
	} else {
		t.SkipNow()
	}
}
