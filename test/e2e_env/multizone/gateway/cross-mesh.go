package gateway

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	k8s_gateway "github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway"
	"github.com/kumahq/kuma/test/e2e_env/multizone/env"
	universal_gateway "github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
)

func MTLSMeshUniversalEgress(name string) InstallFunc {
	mesh := fmt.Sprintf(`
type: Mesh
name: %s
mtls:
  enabledBackend: ca-1
  backends:
    - name: ca-1
      type: builtin
routing:
  zoneEgress: true
`, name)
	return YamlUniversal(mesh)
}

func CrossMeshGatewayOnMultizone() {
	const gatewayClientNamespaceOtherMesh = "cross-mesh-kuma-client-other"
	const gatewayClientNamespaceSameMesh = "cross-mesh-kuma-client"
	const gatewayTestNamespace = "cross-mesh-kuma-test"

	const crossMeshHostname = "gateway.mesh"

	echoServerName := func(mesh string) string {
		return fmt.Sprintf("echo-server-%s", mesh)
	}
	echoServerService := func(mesh, namespace string) string {
		return fmt.Sprintf("%s_%s_svc_80", echoServerName(mesh), namespace)
	}

	const crossMeshGatewayName = "cross-mesh-gateway"
	const edgeGatewayName = "cross-mesh-edge-gateway"

	const gatewayMesh = "cross-mesh-gateway"
	const gatewayOtherMesh = "cross-mesh-other"

	const crossMeshGatewayPort = 9080
	const edgeGatewayPort = 9081

	echoServerApp := func(mesh string) InstallFunc {
		return testserver.Install(
			testserver.WithMesh(mesh),
			testserver.WithName(echoServerName(mesh)),
			testserver.WithNamespace(gatewayTestNamespace),
			testserver.WithEchoArgs("echo", "--instance", mesh),
		)
	}

	crossMeshGatewayYaml := universal_gateway.MkGateway(
		crossMeshGatewayName, gatewayMesh, true, crossMeshHostname, echoServerService(gatewayMesh, gatewayTestNamespace), crossMeshGatewayPort,
	)
	crossMeshGatewayInstanceYaml := k8s_gateway.MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace, gatewayMesh)
	edgeGatewayYaml := universal_gateway.MkGateway(
		edgeGatewayName, gatewayOtherMesh, false, "", echoServerService(gatewayOtherMesh, gatewayTestNamespace), edgeGatewayPort,
	)
	edgeGatewayInstanceYaml := k8s_gateway.MkGatewayInstance(
		edgeGatewayName, gatewayTestNamespace, gatewayOtherMesh,
	)

	BeforeAll(func() {
		globalSetup := NewClusterSetup().
			Install(MTLSMeshUniversalEgress(gatewayMesh)).
			Install(MTLSMeshUniversalEgress(gatewayOtherMesh)).
			Install(YamlUniversal(crossMeshGatewayYaml)).
			Install(YamlUniversal(edgeGatewayYaml))
		Expect(globalSetup.Setup(env.Global)).To(Succeed())

		gatewayZoneSetup := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(echoServerApp(gatewayMesh)).
			Install(echoServerApp(gatewayOtherMesh)).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespaceOtherMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceSameMesh)).
			Install(YamlK8s(crossMeshGatewayInstanceYaml)).
			Install(YamlK8s(edgeGatewayInstanceYaml))
		Expect(gatewayZoneSetup.Setup(env.KubeZone1)).To(Succeed())

		otherZoneSetup := NewClusterSetup().
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(DemoClientK8s(gatewayOtherMesh, gatewayClientNamespaceOtherMesh)).
			Install(DemoClientK8s(gatewayMesh, gatewayClientNamespaceSameMesh))
		Expect(otherZoneSetup.Setup(env.KubeZone2)).To(Succeed())
	})

	E2EAfterAll(func() {
		Expect(env.KubeZone1.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(env.KubeZone1.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())
		Expect(env.KubeZone1.TriggerDeleteNamespace(gatewayTestNamespace)).To(Succeed())

		Expect(env.KubeZone2.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(env.KubeZone2.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())

		Expect(env.Global.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(env.Global.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
		Context("Intrazone", func() {
			zone := &env.KubeZone1
			It("proxies HTTP requests from a different mesh", func() {
				k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceOtherMesh,
				)
			})
			It("proxies HTTP requests from the same mesh", func() {
				k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceSameMesh,
				)
			})
		})
		Context("Crosszone", func() {
			zone := &env.KubeZone2
			filter := fmt.Sprintf(
				"cluster.%s_%s.upstream_rq_total",
				gatewayMesh,
				crossMeshGatewayName,
			)
			var currentStat stats.StatItem

			BeforeEach(func() {
				egress := (*zone).GetZoneEgressEnvoyTunnel()
				stats, err := egress.GetStats(filter)
				Expect(err).ToNot(HaveOccurred())
				Expect(stats.Stats).To(HaveLen(1))
				currentStat = stats.Stats[0]
			})

			It("proxies HTTP requests from a different mesh", func() {
				k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceOtherMesh,
				)

				egress := (*zone).GetZoneEgressEnvoyTunnel()
				Expect(egress.GetStats(filter)).To(stats.BeGreaterThan(currentStat))
			})
			It("proxies HTTP requests from the same mesh", func() {
				k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceSameMesh,
				)

				egress := (*zone).GetZoneEgressEnvoyTunnel()
				Expect(egress.GetStats(filter)).To(stats.BeGreaterThan(currentStat))
			})
		})
	})
}
