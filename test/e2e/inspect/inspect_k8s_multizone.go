package inspect

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func KubernetesMultizone() {
	var globalK8s, zoneK8s *K8sCluster
	var zoneIngress *kube_core.Pod

	meshMTLSOn := func(mesh string) string {
		return fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: Mesh
metadata:
  name: %s
spec:
  mtls:
    enabledBackend: ca-1
    backends:
    - name: ca-1
      type: builtin
`, mesh)
	}

	GetPod := func(namespace, app string) *kube_core.Pod {
		pods, err := k8s.ListPodsE(
			zoneK8s.GetTesting(),
			zoneK8s.GetKubectlOptions(namespace),
			metav1.ListOptions{
				LabelSelector: fmt.Sprintf("app=%s", app),
			},
		)
		Expect(err).ToNot(HaveOccurred())
		Expect(pods).To(HaveLen(1))

		return &pods[0]
	}

	BeforeEach(func() {
		k8sClusters, err := NewK8sClusters([]string{Kuma1, Kuma2}, Silent)
		Expect(err).ToNot(HaveOccurred())

		globalK8s = k8sClusters.GetCluster(Kuma1).(*K8sCluster)
		zoneK8s = k8sClusters.GetCluster(Kuma2).(*K8sCluster)

		err = NewClusterSetup().
			Install(Kuma(config_core.Global)).
			Install(YamlK8s(meshMTLSOn("default"))).
			Setup(globalK8s)
		Expect(err).ToNot(HaveOccurred())

		err = NewClusterSetup().
			Install(Kuma(config_core.Zone,
				WithIngress(),
				WithGlobalAddress(globalK8s.GetKuma().GetKDSServerAddress()),
			)).
			Install(NamespaceWithSidecarInjection(TestNamespace)).
			Install(DemoClientK8s("default")).
			Setup(zoneK8s)
		Expect(err).ToNot(HaveOccurred())
	})

	// Before each test, verify that we have the Dataplanes that we expect to need.
	JustBeforeEach(func() {
		Expect(globalK8s.VerifyKuma()).To(Succeed())
		Expect(zoneK8s.VerifyKuma()).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := globalK8s.GetKumactlOptions().KumactlList("dataplanes", "default")
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("demo-client")))
		}, "60s", "1s").Should(Succeed())

		zoneIngress = GetPod(Config.KumaNamespace, "kuma-ingress")
	})

	E2EAfterEach(func() {
		Expect(zoneK8s.DeleteNamespace(TestNamespace)).To(Succeed())
		Expect(zoneK8s.DeleteKuma()).To(Succeed())
		Expect(zoneK8s.DismissCluster()).To(Succeed())

		Expect(globalK8s.DeleteKuma()).To(Succeed())
		Expect(globalK8s.DismissCluster()).To(Succeed())
	})

	It("should return envoy config_dump for zone ingress", func() {
		zoneIngressName := fmt.Sprintf("%s.%s", zoneIngress.GetName(), Config.KumaNamespace)
		Eventually(func(g Gomega) {
			stdout, err := zoneK8s.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zoneingress", zoneIngressName, "--config-dump")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(stdout).To(ContainSubstring(`"dataplane.proxyType": "ingress"`))

			// filterChainMatches could be available not immediately
			g.Expect(stdout).To(ContainSubstring(`"demo-client_kuma-test_svc{mesh=default}"`))
		}, "30s", "1s").Should(Succeed())
	})
}
