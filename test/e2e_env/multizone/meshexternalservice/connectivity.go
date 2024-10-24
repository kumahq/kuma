package meshexternalservice

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/resources/samples"
	"github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
)

func MesConnectivity() {
	namespace := "mesconnectivity"
	clientNamespace := "mesconnectivity-client"
	meshName := "mesconnectivity"

	esHttpName := "mes-http"
	var esHttpContainerName string

	filter := fmt.Sprintf(
		"cluster.%s_%s__kuma-4_extsvc_80.upstream_rq_total",
		meshName,
		"mesconnectivity-universal-uni1",
	)

	meshGateway := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshGateway
metadata:
  name: edge-gateway-mes
  labels:
    kuma.io/origin: zone
mesh: %s
spec:
  selectors:
  - match:
      kuma.io/service: edge-gateway-mes_%s_svc
  conf:
    listeners:
    - port: 8080
      protocol: HTTP
`, meshName, namespace)
	gatewayRoute := fmt.Sprintf(`
apiVersion: kuma.io/v1alpha1
kind: MeshHTTPRoute
metadata:
  name: route-mes-connectivity
  namespace: %s
  labels:
    kuma.io/mesh: %s
    kuma.io/origin: zone
spec:
  targetRef:
    kind: MeshGateway
    name: edge-gateway-mes
  to:
    - targetRef:
        kind: Mesh
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: /mes
          default:
            backendRefs:
              - kind: MeshExternalService
                labels:
                  kuma.io/display-name: mesconnectivity-universal-uni1
                port: 80
                weight: 1
`, Config.KumaNamespace, meshName)

	BeforeAll(func() {
		esHttpContainerName = fmt.Sprintf("%s_%s_%s", multizone.UniZone1.Name(), meshName, esHttpName)

		// global
		err := NewClusterSetup().
			Install(Yaml(samples.MeshMTLSBuilder().
				WithName(meshName).
				WithoutPassthrough().
				WithEgressRoutingEnabled(),
			)).
			Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
			Setup(multizone.Global)
		Expect(err).ToNot(HaveOccurred())

		// kuma-1
		err = NewClusterSetup().
			Install(NamespaceWithSidecarInjection(namespace)).
			Install(Namespace(clientNamespace)).
			Install(YamlK8s(gatewayRoute)).
			Install(YamlK8s(meshGateway)).
			Install(YamlK8s(gateway.MkGatewayInstance("edge-gateway-mes", namespace, meshName))).
			Install(democlient.Install(
				democlient.WithName("demo-client-gateway"),
				democlient.WithNamespace(clientNamespace),
			)).
			Install(democlient.Install(democlient.WithNamespace(namespace), democlient.WithMesh(meshName))).
			Setup(multizone.KubeZone1)
		Expect(err).ToNot(HaveOccurred())

		// kuma-4
		err = NewClusterSetup().
			Install(DemoClientUniversal("uni-demo-client", meshName, WithTransparentProxy(true))).
			Install(TestServerExternalServiceUniversal(esHttpName, 80, false, WithDockerContainerName(esHttpContainerName))).
			Install(YamlUniversal(fmt.Sprintf(`
type: MeshExternalService
name: mesconnectivity-universal-uni1
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: %s
      port: 80
`, meshName, esHttpContainerName))).
			Setup(multizone.UniZone1)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, meshName)
		DebugUniversal(multizone.UniZone1, meshName)
		DebugKube(multizone.KubeZone1, meshName, namespace)
	})

	BeforeEach(func() {
		Expect(multizone.KubeZone1.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
		Expect(multizone.UniZone1.GetZoneIngressEnvoyTunnel().ResetCounters()).To(Succeed())
		Expect(multizone.UniZone1.GetZoneEgressEnvoyTunnel().ResetCounters()).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(namespace)).To(Succeed())
		Expect(multizone.UniZone1.DeleteMeshApps(meshName)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(meshName)).To(Succeed())
	})

	It("should route from kube zone through universal zone to MeshExternalService", func() {
		// cannot reach from kuma-1 by local hostname generator to Mes in kuma-4
		Consistently(func(g Gomega) {
			// when
			response, err := client.CollectFailure(
				multizone.KubeZone1, "demo-client", "mesconnectivity-universal-uni1.extsvc.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Exitcode).To(Or(Equal(6), Equal(28)))
		}, "5s", "1s").Should(Succeed())

		// can reach Mes in local kuma-4 zone
		Eventually(func(g Gomega) {
			// when
			response, err := client.CollectEchoResponse(
				multizone.UniZone1, "uni-demo-client", "mesconnectivity-universal-uni1.extsvc.mesh.local",
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("mes-http"))
		}).Should(Succeed())

		// can reach by synced resource hostname generator
		Eventually(func(g Gomega) {
			// when
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client", "mesconnectivity-universal-uni1.extsvc.kuma-4.mesh.local",
				client.FromKubernetesPod(namespace, "demo-client"),
			)

			// then
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("mes-http"))
		}).Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(multizone.KubeZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).To(stats.BeGreaterThanZero())
		}, "5s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(multizone.UniZone1.GetZoneIngressEnvoyTunnel().GetStats(filter)).To(stats.BeGreaterThanZero())
		}, "5s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(multizone.UniZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).To(stats.BeGreaterThanZero())
		}, "5s", "1s").Should(Succeed())
	})

	It("should route from gateway in zone kuma-1 to MeshExternalService through zone kuma-4", func() {
		Eventually(func(g Gomega) {
			response, err := client.CollectEchoResponse(
				multizone.KubeZone1, "demo-client-gateway",
				fmt.Sprintf("http://edge-gateway-mes.%s:8080/mes", namespace),
				client.FromKubernetesPod(clientNamespace, "demo-client-gateway"),
			)

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(response.Instance).To(Equal("mes-http"))
		}, "30s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(multizone.KubeZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).To(stats.BeGreaterThanZero())
		}, "5s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(multizone.UniZone1.GetZoneIngressEnvoyTunnel().GetStats(filter)).To(stats.BeGreaterThanZero())
		}, "5s", "1s").Should(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(multizone.UniZone1.GetZoneEgressEnvoyTunnel().GetStats(filter)).To(stats.BeGreaterThanZero())
		}, "5s", "1s").Should(Succeed())
	})
}
