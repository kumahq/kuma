package meshtls

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	meshtls_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	framework_client "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshTLS() {
	const meshName = "multizone-meshtls"
	const k8sZoneNamespace = "multizone-meshtls"

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(meshName)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(meshName, multizone.Zones())).To(Succeed())

		group := errgroup.Group{}
		// Kube Zone 1
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(k8sZoneNamespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(meshName),
				testserver.WithNamespace(k8sZoneNamespace),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-1"),
			)).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(k8sZoneNamespace)).
			Install(democlient.Install(
				democlient.WithName("demo-client"),
				democlient.WithMesh(meshName),
				democlient.WithNamespace(k8sZoneNamespace),
			)).
			SetupInGroup(multizone.KubeZone2, &group)
		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugKube(multizone.KubeZone1, meshName, k8sZoneNamespace)
		DebugKube(multizone.KubeZone2, meshName, k8sZoneNamespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.Global, meshName, meshtls_api.MeshTLSResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(k8sZoneNamespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(k8sZoneNamespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should define TLS version and traffic should works", func() {
		policy := fmt.Sprintf(`
type: MeshTLS
mesh: %s
name: mesh-tls-policy
spec:
  targetRef:
    kind: Mesh
  from:
    - targetRef:
        kind: Mesh
      default:
        tlsVersion:
          min: TLS13
          max: TLS13`, meshName)

		Eventually(func(g Gomega) {
			resp, err := framework_client.CollectEchoResponse(
				multizone.KubeZone2, "demo-client", "test-server_multizone-meshtls_svc_80.mesh",
				framework_client.FromKubernetesPod(k8sZoneNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-1"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())

		// when
		Expect(multizone.Global.Install(YamlUniversal(policy))).To(Succeed())

		// then
		// traffic should still works
		Eventually(func(g Gomega) {
			resp, err := framework_client.CollectEchoResponse(
				multizone.KubeZone2, "demo-client", "test-server_multizone-meshtls_svc_80.mesh",
				framework_client.FromKubernetesPod(k8sZoneNamespace, "demo-client"),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(resp.Instance).To(Equal("kube-test-server-1"))
		}, "30s", "1s").MustPassRepeatedly(5).Should(Succeed())
	})
}
