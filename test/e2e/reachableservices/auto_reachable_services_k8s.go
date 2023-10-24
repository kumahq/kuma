package reachableservices

import (
	"fmt"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var k8sCluster Cluster

var _ = E2EBeforeSuite(func() {
	k8sCluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

	err := NewClusterSetup().
		Install(Kuma(config_core.Standalone,
			WithEnv("KUMA_EXPERIMENTAL_AUTO_REACHABLE_SERVICES", "true"),
		)).
		Install(NamespaceWithSidecarInjection(TestNamespace)).
		Install(testserver.Install(testserver.WithName("client-server"), testserver.WithMesh("default"))).
		Install(testserver.Install(testserver.WithName("first-test-server"), testserver.WithMesh("default"))).
		Install(testserver.Install(testserver.WithName("second-test-server"), testserver.WithMesh("default"))).
		Setup(k8sCluster)

	Expect(err).ToNot(HaveOccurred())

	E2EDeferCleanup(func() {
		Expect(k8sCluster.DeleteKuma()).To(Succeed())
		Expect(k8sCluster.DismissCluster()).To(Succeed())
	})
})

func AutoReachableServices() {
	It("should not connect to non auto reachable service", func() {
		// when
		Expect(YamlK8s(fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshTrafficPermission
metadata:
  name: mtp1
  namespace: %s
  labels:
    kuma.io/mesh: default
spec:
  targetRef:
    kind: MeshService
    name: first-test-server_kuma-test_svc_80
  from:
    - targetRef:
        kind: Mesh
      default:
        action: Deny
    - targetRef:
        kind: MeshService
        name: client-server_kuma-test_svc_80
      default:
        action: Allow
`, Config.KumaNamespace))(k8sCluster)).To(Succeed())

		// then
		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(k8sCluster, "second-test-server", TestNamespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := k8sCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod + "." + TestNamespace, "--type=clusters")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(Not(ContainSubstring("first-test-server_kuma-test_svc_80")))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			pod, err := PodNameOfApp(k8sCluster, "client-server", TestNamespace)
			g.Expect(err).ToNot(HaveOccurred())
			stdout, err := k8sCluster.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "dataplane", pod + "." + TestNamespace, "--type=clusters")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring("first-test-server_kuma-test_svc_80"))
		}, "30s", "1s").Should(Succeed())
	})
}
