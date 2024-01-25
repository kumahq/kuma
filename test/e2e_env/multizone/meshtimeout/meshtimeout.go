package meshtimeout

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshretry_api "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	. "github.com/kumahq/kuma/test/framework"
	framework_client "github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MeshTimeout() {
	const mesh = "mutlizone-meshtimeout"
	const k8sZoneNamespace = "mutlizone-meshtimeout-ns"

	BeforeAll(func() {
		// Global
		Expect(multizone.Global.Install(MeshUniversal(mesh))).To(Succeed())
		Expect(WaitForMesh(mesh, multizone.Zones())).To(Succeed())

		// Kube Zone 1
		Expect(NewClusterSetup().
			Install(NamespaceWithSidecarInjection(k8sZoneNamespace)).
			Install(testserver.Install(
				testserver.WithName("test-client"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(k8sZoneNamespace),
			)).
			Install(testserver.Install(
				testserver.WithName("test-server"),
				testserver.WithMesh(mesh),
				testserver.WithNamespace(k8sZoneNamespace),
			)).
			Setup(multizone.KubeZone1),
		).To(Succeed())

		Expect(DeleteMeshResources(multizone.Global, mesh, meshretry_api.MeshRetryResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(multizone.Global, mesh, meshtimeout_api.MeshTimeoutResourceTypeDescriptor)).To(Succeed())
		Expect(DeleteMeshResources(multizone.Global, mesh, meshhttproute_api.MeshHTTPRouteResourceTypeDescriptor)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(k8sZoneNamespace)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(mesh)).To(Succeed())
	})

	It("should apply MeshTimeout policy to MeshHTTPRoute", func() {
		// when
		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshHTTPRoute
name: route-1
mesh: %s
spec:
  targetRef:
    kind: MeshService
    name: test-client_mutlizone-meshtimeout-ns_svc_80
  to:
    - targetRef:
        kind: MeshService
        name: test-server_mutlizone-meshtimeout-ns_svc_80
      rules: 
        - matches:
            - path: 
                type: PathPrefix
                value: /path/with/timeout
          default:
            backendRefs:
              - kind: MeshService
                name: test-server_mutlizone-meshtimeout-ns_svc_80
                weight: 1
`, mesh))(multizone.Global)).To(Succeed())

		Expect(YamlUniversal(fmt.Sprintf(`
type: MeshTimeout
name: mt1
mesh: %s
spec:
  targetRef:
    kind: MeshHTTPRoute
    name: route-1
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 2s
`, mesh))(multizone.Global)).To(Succeed())

		Eventually(func(g Gomega) {
			start := time.Now()
			_, err := framework_client.CollectEchoResponse(
				multizone.KubeZone1, "test-client", "test-server_mutlizone-meshtimeout-ns_svc_80.mesh/path/without/timeout",
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithHeader("x-set-response-delay-ms", "5000"),
				framework_client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			response, err := framework_client.CollectFailure(
				multizone.KubeZone1, "test-client", "test-server_mutlizone-meshtimeout-ns_svc_80.mesh/path/with/timeout",
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithHeader("x-set-response-delay-ms", "5000"),
				framework_client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "30s", "1s").Should(Succeed())
	})

	It("should apply MeshTimeout policy on Zone CP", func() {
		Eventually(func(g Gomega) {
			start := time.Now()
			_, err := framework_client.CollectEchoResponse(
				multizone.KubeZone1, "test-client", "test-server_mutlizone-meshtimeout-ns_svc_80.mesh",
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithHeader("x-set-response-delay-ms", "5000"),
				framework_client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(time.Since(start)).To(BeNumerically(">", time.Second*5))
		}, "30s", "1s").Should(Succeed())

		Expect(YamlK8s(fmt.Sprintf(`
kind: MeshTimeout
apiVersion: kuma.io/v1alpha1
metadata:
  name: mt-on-zone
  namespace: %s
  labels:
    kuma.io/origin: zone
    kuma.io/mesh: %s
spec:
  targetRef:
    kind: Mesh
  to:
    - targetRef:
        kind: Mesh
      default:
        http:
          requestTimeout: 2s
`, Config.KumaNamespace, mesh))(multizone.KubeZone1)).To(Succeed())

		Eventually(func(g Gomega) {
			response, err := framework_client.CollectFailure(
				multizone.KubeZone1, "test-client", "test-server_mutlizone-meshtimeout-ns_svc_80.mesh",
				framework_client.FromKubernetesPod(k8sZoneNamespace, "test-client"),
				framework_client.WithHeader("x-set-response-delay-ms", "5000"),
				framework_client.WithMaxTime(10),
			)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.ResponseCode).To(Equal(504))
		}, "30s", "1s").Should(Succeed())
	})
}
