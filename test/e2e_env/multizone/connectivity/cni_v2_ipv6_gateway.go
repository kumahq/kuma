package connectivity

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func GatewayIPV6CNIV2() {
	namespace := "gw-ipv6-cniv2"
	meshName := "gw-ipv6-cniv2"

	BeforeAll(func() {
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		err := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(democlient.Install(
				democlient.WithNamespace(namespace),
				democlient.WithMesh(meshName),
				democlient.WithPodAnnotations(map[string]string{
					metadata.KumaGatewayAnnotation: "enabled",
				}),
			)).
			Install(testserver.Install(
				testserver.WithNamespace(namespace),
				testserver.WithMesh(meshName),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server"),
			)).
			Setup(multizone.KubeZone2)
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("client should communicate with server", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(multizone.KubeZone2, "demo-client", "http://test-server_gw-ipv6-cniv2_svc_80.mesh",
				client.FromKubernetesPod(meshName, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("kube-test-server"))
		}, "30s", "1s").Should(Succeed())
	})
}
