package transparentproxy

import (
	"fmt"
	"slices"

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

func TransparentProxyConfigmap() {
	meshName := "transparentproxy-configmap"
	namespace := "transparentproxy-configmap"
	namespaceExternal := "transparentproxy-configmap-external"

	var cluster *K8sCluster
	var demoClientPod kube_core.Pod

	BeforeAll(func() {
		cluster = NewK8sCluster(NewTestingT(), Kuma2, Silent)

		Eventually(func() error {
			return cluster.Install(Kuma(config_core.Standalone, slices.Concat(
				[]KumaDeploymentOption{
					// Occasionally CP will lose a leader in the E2E test just because of this deadline,
					// which does not make sense in such controlled environment (one k3d node, one instance of the CP).
					// 100 s and 80s are values that we also use in mesh-perf when we put a lot of pressure on the CP.
					WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_LEASE_DURATION", "100s"),
					WithEnv("KUMA_RUNTIME_KUBERNETES_LEADER_ELECTION_RENEW_DEADLINE", "80s"),
					WithCtlOpts(map[string]string{
						"--set": fmt.Sprintf("%stransparentProxy.configMap.enabled=true", Config.HelmSubChartPrefix),
					}),
				},
				KumaDeploymentOptionsFromConfig(Config.KumaCpConfig.Standalone.Kubernetes),
			)...))
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
		DebugKube(cluster, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(cluster.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(cluster.TriggerDeleteNamespace(namespaceExternal)).To(Succeed())
		Expect(cluster.DeleteKuma()).To(Succeed())
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
