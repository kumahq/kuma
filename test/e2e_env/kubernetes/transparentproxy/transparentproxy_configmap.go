package transparentproxy

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kube_core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	core_xds "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/kubernetes"
)

func TransparentProxyConfigmap() {
	meshName := "transparentproxy-configmap"
	namespace := "transparentproxy-configmap"
	namespaceExternal := "transparentproxy-configmap-external"

	var demoClientPod kube_core.Pod

	BeforeAll(func() {
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
		Expect(kubernetes.Cluster.TriggerDeleteNamespace(namespaceExternal)).To(Succeed())
		Expect(kubernetes.Cluster.DeleteMesh(meshName)).To(Succeed())
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
			Setup(kubernetes.Cluster)).To(Succeed())
	})
}
