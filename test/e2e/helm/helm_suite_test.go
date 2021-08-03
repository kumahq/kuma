package helm_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2EHelm(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Helm Suite")
	} else {
		t.SkipNow()
	}
}
