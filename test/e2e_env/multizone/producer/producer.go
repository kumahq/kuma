package producer

import (
	"fmt"

	"github.com/gruntwork-io/terratest/modules/k8s"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/kds/hash"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func ProducerPolicyFlow() {
	const mesh = "producer-policy-flow"
	const k8sZoneNamespace = "producer-policy-flow-ns"

	BeforeAll(func() {
		// Global
		Expect(NewClusterSetup().
			Install(MTLSMeshUniversal(mesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
			Setup(multizone.Global)).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Kube Zone 1
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(k8sZoneNamespace)).
			Install(testserver.Install(
				testserver.WithName("test-client"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(k8sZoneNamespace),
			)).
			Setup(multizone.KubeZone1),
		).To(Succeed())

		kubeServiceYAML := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshService
metadata:
  name: test-server
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
    kuma.io/managed-by: k8s-controller
    k8s.kuma.io/is-headless-service: "false"
spec:
  selector:
    dataplaneTags:
      app: test-server
      k8s.kuma.io/namespace: %s
  ports:
  - port: 80
    name: main
    targetPort: main
    appProtocol: http
`, k8sZoneNamespace, mesh, k8sZoneNamespace)

		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(k8sZoneNamespace)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(k8sZoneNamespace),
				testserver.WithEchoArgs("echo", "--instance", "kube-test-server-2"),
			)).
			Install(YamlK8s(kubeServiceYAML)).
			Setup(multizone.KubeZone2),
		).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, mesh)
		DebugKube(multizone.KubeZone1, mesh, k8sZoneNamespace)
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.Global, mesh, meshtimeout_api.MeshTimeoutResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.Global, mesh, meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(k8sZoneNamespace)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(k8sZoneNamespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	It("should synced producer policy to other clusters", func() {
		Expect(YamlK8s(fmt.Sprintf(`
kind: MeshTimeout
apiVersion: kuma.io/v1alpha1
metadata:
  name: to-test-server
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
      default:
        http:
          requestTimeout: 2s
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				multizone.KubeZone1.GetTesting(),
				multizone.KubeZone1.GetKubectlOptions(Config.KumaNamespace),
				"get", "meshtimeouts")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).To(ContainSubstring(hash.HashedName(mesh, "to-test-server", Kuma2, k8sZoneNamespace)))
		}).Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
kind: MeshTimeout
apiVersion: kuma.io/v1alpha1
metadata:
  name: to-test-server
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
    kuma.io/policy-role: consumer
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: MeshService
        name: test-server
        namespace: random-ns-name
      default:
        http:
          requestTimeout: 2s
`, k8sZoneNamespace, mesh))(multizone.KubeZone2)).To(Succeed())

		Eventually(func(g Gomega) {
			out, err := k8s.RunKubectlAndGetOutputE(
				multizone.KubeZone1.GetTesting(),
				multizone.KubeZone1.GetKubectlOptions(Config.KumaNamespace),
				"get", "meshtimeouts")
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(out).ToNot(ContainSubstring(hash.HashedName(mesh, "to-test-server", Kuma2, k8sZoneNamespace)))
		}).Should(Succeed())
	})
}
