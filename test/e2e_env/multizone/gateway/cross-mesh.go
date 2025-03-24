package gateway

import (
	"fmt"
	"net"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/sync/errgroup"

	k8s_gateway "github.com/kumahq/kuma/test/e2e_env/kubernetes/gateway"
	universal_gateway "github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/democlient"
	"github.com/kumahq/kuma/test/framework/deployments/testserver"
	"github.com/kumahq/kuma/test/framework/envoy_admin/stats"
	"github.com/kumahq/kuma/test/framework/envs/multizone"
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

	crossMeshGatewayServiceName := fmt.Sprintf("%s_%s_svc", crossMeshGatewayName, gatewayTestNamespace)
	edgeGatewayServiceName := fmt.Sprintf("%s_%s_svc", edgeGatewayName, gatewayTestNamespace)

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
		crossMeshGatewayName, gatewayMesh, crossMeshGatewayServiceName, true, crossMeshHostname, echoServerService(gatewayMesh, gatewayTestNamespace), crossMeshGatewayPort,
	)
	crossMeshGatewayInstanceYaml := k8s_gateway.MkGatewayInstance(crossMeshGatewayName, gatewayTestNamespace, gatewayMesh)
	edgeGatewayYaml := universal_gateway.MkGateway(
		edgeGatewayName, gatewayOtherMesh, edgeGatewayServiceName, false, "", echoServerService(gatewayOtherMesh, gatewayTestNamespace), edgeGatewayPort,
	)
	edgeGatewayInstanceYaml := k8s_gateway.MkGatewayInstance(
		edgeGatewayName, gatewayTestNamespace, gatewayOtherMesh,
	)

	BeforeAll(func() {
		globalSetup := NewClusterSetup().
			Install(MTLSMeshUniversalEgress(gatewayMesh)).
			Install(MTLSMeshUniversalEgress(gatewayOtherMesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(gatewayMesh)).
			Install(MeshTrafficPermissionAllowAllUniversal(gatewayOtherMesh)).
			Install(YamlUniversal(crossMeshGatewayYaml)).
			Install(YamlUniversal(edgeGatewayYaml))
		Expect(globalSetup.Setup(multizone.Global)).To(Succeed())

		group := errgroup.Group{}
		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(gatewayTestNamespace)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(Parallel(
				echoServerApp(gatewayMesh),
				echoServerApp(gatewayOtherMesh),
				democlient.Install(democlient.WithMesh(gatewayOtherMesh), democlient.WithNamespace(gatewayClientNamespaceOtherMesh)),
				democlient.Install(democlient.WithMesh(gatewayMesh), democlient.WithNamespace(gatewayClientNamespaceSameMesh)),
			)).
			Install(YamlK8s(crossMeshGatewayInstanceYaml)).
			Install(YamlK8s(edgeGatewayInstanceYaml)).
			SetupInGroup(multizone.KubeZone1, &group)

		NewClusterSetup().
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceOtherMesh)).
			Install(NamespaceWithSidecarInjection(gatewayClientNamespaceSameMesh)).
			Install(Parallel(
				democlient.Install(democlient.WithMesh(gatewayOtherMesh), democlient.WithNamespace(gatewayClientNamespaceOtherMesh)),
				democlient.Install(democlient.WithMesh(gatewayMesh), democlient.WithNamespace(gatewayClientNamespaceSameMesh)),
			)).
			SetupInGroup(multizone.KubeZone2, &group)

		Expect(group.Wait()).To(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(multizone.Global, gatewayMesh)
		DebugUniversal(multizone.Global, gatewayOtherMesh)
		DebugKube(multizone.KubeZone1, gatewayMesh, gatewayClientNamespaceOtherMesh, gatewayClientNamespaceSameMesh, gatewayTestNamespace)
		DebugKube(multizone.KubeZone2, gatewayOtherMesh, gatewayClientNamespaceOtherMesh, gatewayClientNamespaceSameMesh)
	})

	E2EAfterAll(func() {
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())
		Expect(multizone.KubeZone1.TriggerDeleteNamespace(gatewayTestNamespace)).To(Succeed())

		Expect(multizone.KubeZone2.TriggerDeleteNamespace(gatewayClientNamespaceOtherMesh)).To(Succeed())
		Expect(multizone.KubeZone2.TriggerDeleteNamespace(gatewayClientNamespaceSameMesh)).To(Succeed())

		Expect(multizone.Global.DeleteMesh(gatewayMesh)).To(Succeed())
		Expect(multizone.Global.DeleteMesh(gatewayOtherMesh)).To(Succeed())
	})

	Context("when mTLS is enabled", func() {
		gatewayAddr := net.JoinHostPort(crossMeshHostname, strconv.Itoa(crossMeshGatewayPort))
		Context("Intrazone", func() {
			zone := &multizone.KubeZone1
			It("proxies HTTP requests from a different mesh", func() {
				Eventually(k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceOtherMesh,
				)).Should(Succeed())
			})
			It("proxies HTTP requests from the same mesh", func() {
				Eventually(k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceSameMesh,
				)).Should(Succeed())
			})
		})
		Context("Crosszone", func() {
			zone := &multizone.KubeZone2
			filter := fmt.Sprintf(
				"cluster.%s_%s.upstream_rq_total",
				gatewayMesh,
				crossMeshGatewayServiceName,
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
				Eventually(k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceOtherMesh,
				)).Should(Succeed())

				egress := (*zone).GetZoneEgressEnvoyTunnel()
				Expect(egress.GetStats(filter)).To(stats.BeGreaterThan(currentStat))
			})
			It("proxies HTTP requests from the same mesh", func() {
				Eventually(k8s_gateway.SuccessfullyProxyRequestToGateway(
					*zone, gatewayMesh,
					gatewayAddr,
					gatewayClientNamespaceSameMesh,
				)).Should(Succeed())

				egress := (*zone).GetZoneEgressEnvoyTunnel()
				Expect(egress.GetStats(filter)).To(stats.BeGreaterThan(currentStat))
			})
		})
	})
}
