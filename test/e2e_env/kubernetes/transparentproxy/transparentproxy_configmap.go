package transparentproxy

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"

	core_xds "github.com/kumahq/kuma/pkg/core/xds/types"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func TransparentProxyConfigmap() {
	meshName := "transparentproxy-configmap"
	namespace := "transparentproxy-configmap"

	var demoClientPod kube_core.Pod

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Parallel(
				democlient.Install(
					democlient.WithMesh(meshName),
					democlient.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
			)).
			Setup(kubernetes.Cluster)).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := kubernetes.Cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("demo-client")))
		}, "10s", "250ms").Should(Succeed())

		var err error
		demoClientPod, err = PodOfApp(kubernetes.Cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(kubernetes.Cluster, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteKuma()).To(Succeed())
	})

	It("should contain transparent proxy in configmap feature flag in the xds metadata", func() {
		stdout, err := kubernetes.Cluster.
			GetKumactlOptions().
			RunKumactlAndGetOutput(
				"inspect",
				"dataplane",
				"--type=config-dump",
				fmt.Sprintf("--mesh=%s", meshName),
				fmt.Sprintf("%s.%s", demoClientPod.Name, namespace),
			)
		Expect(err).ToNot(HaveOccurred())

		Expect(stdout).To(ContainSubstring(core_xds.FeatureTransparentProxyInDataplaneMetadata))
	})

	It("should not contain transparentProxying configuration in the Dataplane object", func() {
		stdout, err := kubernetes.Cluster.
			GetKumactlOptions().
			RunKumactlAndGetOutput(
				"get",
				"dataplane",
				"--output=json",
				fmt.Sprintf("--mesh=%s", meshName),
				fmt.Sprintf("%s.%s", demoClientPod.Name, namespace),
			)
		Expect(err).ToNot(HaveOccurred())

		Expect(stdout).ToNot(ContainSubstring(`"transparentProxying":`))
	})

	It("should be able to connect to test-server", func() {
		Eventually(func(g Gomega) {
			_, err := client.CollectFailure(
				kubernetes.Cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})
}
