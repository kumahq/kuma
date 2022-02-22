package tracing_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/tracing"
)

func TestE2E(t *testing.T) {
	test.RunSpecs(t, "E2E Tracing Suite")
}

var _ = Describe("Tracing K8S", tracing.TracingK8S)
var _ = Describe("Tracing Universal", tracing.TracingUniversal)
