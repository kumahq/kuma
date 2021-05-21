package universal_test

import (
	"testing"

	"github.com/kumahq/kuma/test/framework"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestE2EHealthCheckUniversal(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		RegisterFailHandler(Fail)
		RunSpecs(t, "Health Check Universal Suite")
	} else {
		t.SkipNow()
	}
}
