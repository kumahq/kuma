package tracing_test

import (
	"testing"

	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/tracing"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2ETracing(t *testing.T) {
	if framework.IsK8sClustersStarted() {
		test.RunSpecs(t, "E2E Tracing Suite")
	} else {
		t.SkipNow()
	}
}

var _ = Describe("Tracing K8S", tracing.TracingK8S)
var _ = Describe("Tracing Universal", tracing.TracingUniversal)
