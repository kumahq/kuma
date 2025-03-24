package reachableservices_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/test/e2e/reachableservices"
	"github.com/kumahq/kuma/test/framework"
)

func TestE2E(t *testing.T) {
	test.RunE2ESpecs(t, "E2E Auto Reachable Services Kubernetes Suite")
}

var _ = framework.E2EBeforeSuite(func() {
	reachableservices.KubeCluster = framework.NewK8sCluster(framework.NewTestingT(), framework.Kuma1, framework.Silent)

	err := framework.NewClusterSetup().
		Install(framework.Kuma(config_core.Zone,
			framework.WithEnv("KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES", "true"),
		)).
		Setup(reachableservices.KubeCluster)

	Expect(err).ToNot(HaveOccurred())
})

var _ = framework.E2EAfterSuite(func() {
	Expect(reachableservices.KubeCluster.DeleteKuma()).To(Succeed())
	Expect(reachableservices.KubeCluster.DismissCluster()).To(Succeed())
})

var (
	_ = Describe("Auto Reachable Services on Kubernetes", Label("job-3"), reachableservices.AutoReachableServices, Ordered)
	_ = Describe("Auto Reachable Mesh Services on Kubernetes", Label("job-3"), reachableservices.AutoReachableMeshServices, Ordered)
)
