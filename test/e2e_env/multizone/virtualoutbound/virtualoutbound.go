package virtualoutbound

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func VirtualOutbound() {
	_ = Describe("No Zone Egress", func() {
		virtualOutbound("virtual-outbounds", samples.MeshMTLSBuilder())
	}, Ordered)

	_ = Describe("Zone Egress", func() {
		virtualOutbound("virtual-outbounds-ze", samples.MeshMTLSBuilder().WithEgressRoutingEnabled())
	}, Ordered)
}

func virtualOutbound(meshName string, meshBuilder *builders.MeshBuilder) {
	namespace := meshName

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(ResourceUniversal(meshBuilder.WithName(meshName).Build())).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithStatefulSet(),
				testserver.WithReplicas(2),
			)).
			SetupInGroup(multizone.KubeZone2, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
		DebugKube(multizone.KubeZone2, meshName, namespace)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	BeforeEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, mesh.VirtualOutboundResourceTypeDescriptor)).To(Succeed())
	})

	It("simple virtual outbound", func() {
		virtualOutboundAll := fmt.Sprintf(`
type: VirtualOutbound
mesh: %s
name: instance
selectors:
  - match:
      kuma.io/service: "*"
conf:
  host: "{{.svc}}.foo"
  port: "8080"
  parameters:
  - name: "svc"
    tagKey: "kuma.io/service"
`, meshName)
		err := multizone.Global.Install(YamlUniversal(virtualOutboundAll))
		Expect(err).ToNot(HaveOccurred())

		// Succeed with virtual-outbound
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", fmt.Sprintf("test-server_%s_svc_80.foo:8080", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(ContainSubstring("test-server"))
		}, "30s", "1s").Should(Succeed())
	})

	It("virtual outbounds on statefulSet", func() {
		virtualOutboundAll := fmt.Sprintf(`
type: VirtualOutbound
mesh: %s
name: statefulsets
selectors:
- match:
    kuma.io/service: "*"
    statefulset.kubernetes.io/pod-name: "*"
conf:
  host: "{{.svc}}.{{.inst}}"
  port: "8080"
  parameters:
  - name: "svc"
    tagKey: "kuma.io/service"
  - name: "inst"
    tagKey: "statefulset.kubernetes.io/pod-name"
`, meshName)
		err := multizone.Global.Install(YamlUniversal(virtualOutboundAll))
		Expect(err).ToNot(HaveOccurred())
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", fmt.Sprintf("test-server_%s_svc_80.test-server-0:8080", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("test-server-0"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", fmt.Sprintf("test-server_%s_svc_80.test-server-1:8080", namespace),
				client.FromKubernetesPod(namespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(ContainSubstring("test-server-1"))
		}, "30s", "1s").Should(Succeed())
	})
}
