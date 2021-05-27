package kubernetes_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/kic/kubernetes"
)

var _ = Describe("Kong Ingress on Kubernetes", kubernetes.KICKubernetes)
