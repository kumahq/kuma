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
	meshproxypatch "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshproxypatch/api/v1alpha1"
	meshratelimit "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	meshretry "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtls "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtls/api/v1alpha1"
	meshtrace "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/api/v1alpha1"
	meshtrafficpermission "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/client"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

const zoneProxyMeshName = "envoyconfig-zoneproxies"

const (
	zoneProxyIngressDP = "zone-proxy-ingress"
	zoneProxyEgressDP  = "zone-proxy-egress"
)

// zoneProxyDpEnvs pins the kuma-dp socket directory to /tmp. Without this,
// kuma-dp creates a randomized /tmp/kuma-dp-<N>/ directory each run and that
// random suffix would leak into the generated socket paths in the goldens,
// making the test flaky.
var dppEnvs = map[string]string{
	"KUMA_DATAPLANE_RUNTIME_SOCKET_DIR":   "/tmp",
	"KUMA_DATAPLANE_RUNTIME_IPV6_ENABLED": "false",
}

func ZoneProxies() {
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
		meshtrace.MeshTraceResourceTypeDescriptor,
		meshmetric.MeshMetricResourceTypeDescriptor,
		meshproxypatch.MeshProxyPatchResourceTypeDescriptor,
		meshtrafficpermission.MeshTrafficPermissionResourceTypeDescriptor,
	))

	DescribeTable("should generate proper Envoy config for zone proxies",
		TestZoneProxyConfig,
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshtrace"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshmetric"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshproxypatch"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshcircuitbreaker"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshhealthcheck"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshtrafficpermission"), "envoyconfig"),
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

	// Use an explicit timeout long enough for a freshly-registered zone-proxy
	// DPP to publish a DataplaneInsight, which kumactl `inspect --shadow`
	// requires. Default Eventually timeout (30s) is sometimes too tight when
	// the inspect runs immediately after the DPP container is registered.
	// then
	Eventually(func(g Gomega) {
		g.Expect(getConfig(zoneProxyMeshName, "zone-proxy-demo-client")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "zone-proxy-demo-client.golden.json", 1)))
		g.Expect(getConfig(zoneProxyMeshName, "zone-proxy-test-server")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "zone-proxy-test-server.golden.json", 1)))
		g.Expect(getConfig(zoneProxyMeshName, zoneProxyIngressDP)).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", zoneProxyIngressDP+".golden.json", 1)))
		g.Expect(getConfig(zoneProxyMeshName, zoneProxyEgressDP)).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", zoneProxyEgressDP+".golden.json", 1)))
	}, "90s", "2s").Should(Succeed())
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
			zoneproxy.WithWorkload(zoneProxyIngressDP),
			zoneproxy.WithDpEnvs(dppEnvs),
		)).
		Install(zoneproxy.Install(
			zoneproxy.WithMesh(zoneProxyMeshName),
			zoneproxy.WithEgressPort(11002),
			zoneproxy.WithWorkload(zoneProxyEgressDP),
			zoneproxy.WithDpEnvs(dppEnvs),
		)).
		Install(DemoClientUniversal("zone-proxy-demo-client", zoneProxyMeshName,
			WithTransparentProxy(true),
			WithWorkload("zone-proxy-demo-client"),
			WithDpEnvs(dppEnvs)),
		).
		Install(TestServerUniversal("zone-proxy-test-server", zoneProxyMeshName,
			WithArgs([]string{"echo", "--instance", "universal-1"}),
			WithServiceName("zone-proxy-test-server"),
			WithWorkload("zone-proxy-test-server"),
			WithDpEnvs(dppEnvs)),
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

	// Wait until both zone-proxy DPPs have a usable DataplaneInsight. Without
	// the insight, `kumactl inspect dataplane --shadow` returns 404 and the
	// table tests would only see that error during their (shorter) polling
	// window. Doing the wait here once amortizes the cost over all entries.
	WaitDataplaneInspectable(universal.Cluster, zoneProxyMeshName, zoneProxyIngressDP)
	WaitDataplaneInspectable(universal.Cluster, zoneProxyMeshName, zoneProxyEgressDP)

	waitMeshServiceReady(zoneProxyMeshName, "zone-proxy-demo-client", fmt.Sprintf("spiffe://%s.%s.mesh.local/workload/zone-proxy-demo-client", zoneProxyMeshName, universal.Cluster.ZoneName()))
	waitMeshServiceReady(zoneProxyMeshName, "zone-proxy-test-server", fmt.Sprintf("spiffe://%s.%s.mesh.local/workload/zone-proxy-test-server", zoneProxyMeshName, universal.Cluster.ZoneName()))

	Eventually(func(g Gomega) {
		_, err := client.CollectEchoResponse(universal.Cluster, "zone-proxy-demo-client", "zone-proxy-test-server.svc.mesh.local")
		g.Expect(err).ToNot(HaveOccurred())
	}).Should(Succeed())
}

func CleanupAfterZoneProxyTest(policies ...core_model.ResourceTypeDescriptor) func() {
	return cleanupAfterTest(zoneProxyMeshName, []string{zoneProxyIngressDP, zoneProxyEgressDP}, policies...)
}

func CleanupAfterZoneProxySuite() {
	Expect(universal.Cluster.DeleteMeshApps(zoneProxyMeshName)).To(Succeed())
	Expect(universal.Cluster.DeleteMesh(zoneProxyMeshName)).To(Succeed())
}

func AfterZoneProxyFailure() {
	DebugUniversal(universal.Cluster, zoneProxyMeshName)
}
