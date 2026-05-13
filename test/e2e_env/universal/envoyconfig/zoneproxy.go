package envoyconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	meshaccesslog "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshcircuitbreaker "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshfaultinjection "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	meshhealthcheck "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	meshhttproute "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshloadbalancing "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	meshmetric "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshratelimit "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	meshretry "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtls "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtls/api/v1alpha1"
	meshtrafficpermission "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

const zoneProxyMeshName = "zoneproxy-envoyconfig"

func ZoneProxy() {
	BeforeAll(SetupZoneProxyCluster)

	AfterEachFailure(AfterZoneProxyFailure)

	E2EAfterAll(CleanupAfterZoneProxySuite)

	E2EAfterEach(CleanupAfterZoneProxyTest(
		meshtimeout.MeshTimeoutResourceTypeDescriptor,
		meshaccesslog.MeshAccessLogResourceTypeDescriptor,
		meshfaultinjection.MeshFaultInjectionResourceTypeDescriptor,
		meshratelimit.MeshRateLimitResourceTypeDescriptor,
		meshtls.MeshTLSResourceTypeDescriptor,
		meshhealthcheck.MeshHealthCheckResourceTypeDescriptor,
		meshcircuitbreaker.MeshCircuitBreakerResourceTypeDescriptor,
		meshhttproute.MeshHTTPRouteResourceTypeDescriptor,
		meshretry.MeshRetryResourceTypeDescriptor,
		meshloadbalancing.MeshLoadBalancingStrategyResourceTypeDescriptor,
		meshmetric.MeshMetricResourceTypeDescriptor,
		meshtrafficpermission.MeshTrafficPermissionResourceTypeDescriptor,
	))

	DescribeTable("should generate proper Envoy config",
		TestZoneProxyConfig,
		test.EntriesForFolder(filepath.Join("zoneproxy", "meshcircuitbreaker"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("zoneproxy", "meshhealthcheck"), "envoyconfig"),
	)
}

func TestZoneProxyConfig(inputFile string) {
	// given
	input, err := os.ReadFile(inputFile)
	Expect(err).ToNot(HaveOccurred())

	// when
	if len(input) > 0 {
		Expect(universal.Cluster.Install(YamlUniversal(string(input)))).To(Succeed())
	}

	// time.Sleep(1 * time.Hour)
	// then
	Eventually(func(g Gomega) {
		g.Expect(getConfig(zoneProxyMeshName, "zone-proxy-demo-client")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "zone-proxy-demo-client.golden.json", 1)))
		g.Expect(getConfig(zoneProxyMeshName, "zone-proxy-test-server")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "zone-proxytest-server.golden.json", 1)))
		g.Expect(getConfig(zoneProxyMeshName, "zone-proxy-ingress")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "zone-proxy-ingress.golden.json", 1)))
		g.Expect(getConfig(zoneProxyMeshName, "zone-proxy-egress")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "zone-proxy-egress.golden.json", 1)))
	}).Should(Succeed())
}

func SetupZoneProxyCluster() {
	meshExternalService := fmt.Sprintf(`
type: MeshExternalService
name: mes-zone-proxy
mesh: %s
labels:
  kuma.io/origin: zone
  kuma.io/access: external
  kuma.io/display-name: mes-zone-proxy
spec:
  match:
    type: HostnameGenerator
    port: 80
    protocol: http
  endpoints:
    - address: 127.0.0.1
      port: 80
`, zoneProxyMeshName)

	meshIdentityYAML := fmt.Sprintf(`
type: MeshIdentity
name: identity
mesh: %s
spec:
  selector:
    dataplane:
      matchLabels: {}
  spiffeID:
    trustDomain: "{{ .Mesh }}.{{ .Zone }}.mesh.local"
  provider:
    type: Bundled
    bundled:
      meshTrustCreation: Enabled
      insecureAllowSelfSigned: true
      certificateParameters:
        expiry: 24h
      autogenerate:
        enabled: true
`, zoneProxyMeshName)

	err := NewClusterSetup().
		Install(
			Yaml(
				builders.Mesh().
					WithName(zoneProxyMeshName).
					WithoutInitialPolicies().
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
			),
		).
		Install(MeshTrafficPermissionAllowAllUniversal(zoneProxyMeshName)).
		Install(YamlUniversal(meshExternalService)).
		Install(YamlUniversal(meshIdentityYAML)).
		Install(zoneproxy.Install(
			zoneproxy.WithMesh(zoneProxyMeshName),
			zoneproxy.WithIngressPort(11001),
			zoneproxy.WithWorkload("zone-proxy-ingress"),
		)).
		Install(zoneproxy.Install(
			zoneproxy.WithMesh(zoneProxyMeshName),
			zoneproxy.WithEgressPort(11002),
			zoneproxy.WithWorkload("zone-proxy-egress"),
		)).
		Install(DemoClientUniversal("zone-proxy-demo-client", zoneProxyMeshName,
			WithTransparentProxy(true),
			WithWorkload("zone-proxy-demo-client"),
			WithDpEnvs(map[string]string{
				"KUMA_DATAPLANE_RUNTIME_SOCKET_DIR":   "/tmp",
				"KUMA_DATAPLANE_RUNTIME_IPV6_ENABLED": "false",
			})),
		).
		Install(TestServerUniversal("zone-proxy-test-server", zoneProxyMeshName,
			WithArgs([]string{"echo", "--instance", "universal-1"}),
			WithServiceName("zone-proxy-test-server"),
			WithWorkload("zone-proxy-test-server"),
			WithDpEnvs(map[string]string{
				"KUMA_DATAPLANE_RUNTIME_SOCKET_DIR":   "/tmp",
				"KUMA_DATAPLANE_RUNTIME_IPV6_ENABLED": "false",
			}),
		),
		).
		Setup(universal.Cluster)
	Expect(err).ToNot(HaveOccurred())

	ingressIP := universal.Cluster.GetApp("zone-proxy-ingress").GetIP()
	Expect(NewClusterSetup().
		Install(YamlUniversal(fmt.Sprintf(`
type: MeshZoneAddress
name: zone-proxy-ingress
mesh: %s
labels:
  kuma.io/origin: zone
spec:
  address: %s
  port: %d
`, zoneProxyMeshName, ingressIP, 11001))).
		Setup(universal.Cluster)).To(Succeed())

	waitMeshServiceReady(zoneProxyMeshName, "zone-proxy-demo-client", fmt.Sprintf("spiffe://%s.%s.mesh.local/workload/zone-proxy-demo-client", zoneProxyMeshName, universal.Cluster.ZoneName()))
	waitMeshServiceReady(zoneProxyMeshName, "zone-proxy-test-server", fmt.Sprintf("spiffe://%s.%s.mesh.local/workload/zone-proxy-test-server", zoneProxyMeshName, universal.Cluster.ZoneName()))

	Eventually(func(g Gomega) {
		_, err := client.CollectEchoResponse(universal.Cluster, "zone-proxy-demo-client", "zone-proxy-test-server.svc.mesh.local")
		g.Expect(err).ToNot(HaveOccurred())
	}).Should(Succeed())
}

func CleanupAfterZoneProxyTest(policies ...core_model.ResourceTypeDescriptor) func() {
	return cleanupAfterTest(zoneProxyMeshName, []string{"zone-proxy-demo-client", "zone-proxy-test-server", "zone-proxy-ingress", "zone-proxy-egress"}, policies...)
}

func CleanupAfterZoneProxySuite() {
	Expect(universal.Cluster.DeleteMeshApps(zoneProxyMeshName)).To(Succeed())
	Expect(universal.Cluster.DeleteMesh(zoneProxyMeshName)).To(Succeed())
}

func AfterZoneProxyFailure() {
	DebugUniversal(universal.Cluster, zoneProxyMeshName)
}
