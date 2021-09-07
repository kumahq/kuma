package kubernetes_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/healthcheck/kubernetes"
)

var _ = Describe("Test Virtual Probes on Kubernetes", kubernetes.VirtualProbes)
