package tracing_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/tracing"
)

var _ = Describe("Tracing K8S", tracing.TracingK8S)
