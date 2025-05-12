package transparentproxy

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	config_core "github.com/kumahq/kuma/pkg/config/core"
	core_xds "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
)

func TransparentProxyConfigMap() {
	const meshName = "transparentproxy-configmap"
	const namespace = "transparentproxy-configmap"
	const namespaceExternal = "transparentproxy-configmap-external"

	var cluster *K8sCluster
	var demoClientPod kube_core.Pod

	BeforeAll(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma1, Silent)

		Eventually(func() error {
			return cluster.Install(Kuma(
				config_core.Zone,
				WithCtlOpts(map[string]string{
					"--set": fmt.Sprintf("%stransparentProxy.configMap.enabled=true", Config.HelmSubChartPrefix),
				}),
			))
		}, "90s", "3s").Should(Succeed())

		Expect(NewClusterSetup().
			Install(MTLSMeshKubernetes(meshName)).
			Install(MeshTrafficPermissionAllowAllKubernetes(meshName)).
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(namespaceExternal)).
			Install(Parallel(
				democlient.Install(
					democlient.WithMesh(meshName),
					democlient.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithMesh(meshName),
					testserver.WithNamespace(namespace),
				),
				testserver.Install(
					testserver.WithNamespace(namespaceExternal),
				),
			)).
			Setup(cluster)).To(Succeed())

		// Synchronize on the dataplanes coming up.
		Eventually(func(g Gomega) {
			dataplanes, err := cluster.GetKumactlOptions().KumactlList("dataplanes", meshName)
			g.Expect(err).ToNot(HaveOccurred())
			// Dataplane names are generated, so we check for a partial match.
			g.Expect(dataplanes).Should(ContainElement(ContainSubstring("demo-client")))
		}, "10s", "250ms").Should(Succeed())

		var err error
		demoClientPod, err = PodOfApp(cluster, "demo-client", namespace)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugKube(cluster, meshName, namespace, namespaceExternal)
	})

	E2EAfterAll(func() {
		Expect(cluster.TriggerDeleteNamespace(namespaceExternal)).To(Succeed())
		Expect(cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
		Expect(cluster.DismissCluster()).To(Succeed())
	})

	It("should contain transparent proxy in configmap feature flag in the xds metadata", func() {
		stdout, err := cluster.
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
		stdout, err := cluster.
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
				cluster,
				"demo-client",
				fmt.Sprintf("test-server_%s_svc_80.mesh", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
		}, "30s", "1s").Should(Succeed())
	})

	It("should be able to use network from init container if we ignore ports for uid", func() {
		tpConfigMapName := "tproxy-config-custom"

		Expect(NewClusterSetup().
			Install(ConfigMapTProxyKubernetes(
				tpConfigMapName,
				namespace,
				`
redirect:
  outbound:
    excludePortsForUIDs:
    - tcp:80:1234
    - udp:53:1234
`,
			)).
			Install(testserver.Install(
				testserver.WithName("another-test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(namespace),
				testserver.WithPodAnnotations(map[string]string{
					metadata.KumaInitFirst:                            metadata.AnnotationTrue,
					metadata.KumaTrafficTransparentProxyConfigMapName: tpConfigMapName,
				}),
				testserver.AddInitContainer(kube_core.Container{
					Name:            "init-test-server",
					Image:           Config.GetUniversalImage(),
					ImagePullPolicy: "IfNotPresent",
					Command:         []string{"curl"},
					Args: []string{
						"--verbose",
						"--max-time", "3",
						"--fail",
						fmt.Sprintf("test-server.%s.svc.cluster.local:80", namespaceExternal),
					},
					Resources: kube_core.ResourceRequirements{
						Limits: kube_core.ResourceList{
							"cpu":    resource.MustParse("50m"),
							"memory": resource.MustParse("64Mi"),
						},
					},
					SecurityContext: &kube_core.SecurityContext{
						RunAsUser: pointer.To(int64(1234)),
					},
				}),
			)).
			Setup(cluster)).To(Succeed())
	})
}
