package envoyconfig

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	meshmetric "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshmetric/api/v1alpha1"
	meshtrace "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrace/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/v2/test/framework"
	"github.com/kumahq/kuma/v2/test/framework/deployments/zoneproxy"
	"github.com/kumahq/kuma/v2/test/framework/envs/universal"
)

const zoneProxyMesh = "envoyconfig-zoneproxies"

const (
	zoneProxyIngressDP = "zone-proxy-ingress"
	zoneProxyEgressDP  = "zone-proxy-egress"
)

// zoneProxyDpEnvs pins the kuma-dp socket directory to /tmp. Without this,
// kuma-dp creates a randomized /tmp/kuma-dp-<N>/ directory each run and that
// random suffix would leak into the generated socket paths in the goldens,
// making the test flaky.
var zoneProxyDpEnvs = map[string]string{
	"KUMA_DATAPLANE_RUNTIME_SOCKET_DIR":   "/tmp",
	"KUMA_DATAPLANE_RUNTIME_IPV6_ENABLED": "false",
}

func ZoneProxies() {
	BeforeAll(SetupZoneProxyCluster)

	AfterEachFailure(AfterZoneProxyFailure)

	E2EAfterAll(CleanupAfterZoneProxySuite)

	E2EAfterEach(CleanupAfterZoneProxyTest(
		meshtrace.MeshTraceResourceTypeDescriptor,
		meshmetric.MeshMetricResourceTypeDescriptor,
	))

	DescribeTable("should generate proper Envoy config for zone proxies",
		TestZoneProxyConfig,
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshtrace"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("zoneproxies", "meshmetric"), "envoyconfig"),
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
	Eventually(func(g Gomega) {
		g.Expect(getConfig(zoneProxyMesh, zoneProxyIngressDP)).
			To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", zoneProxyIngressDP+".golden.json", 1)))
		g.Expect(getConfig(zoneProxyMesh, zoneProxyEgressDP)).
			To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", zoneProxyEgressDP+".golden.json", 1)))
	}, "90s", "2s").Should(Succeed())
}

func SetupZoneProxyCluster() {
	err := NewClusterSetup().
		Install(
			Yaml(
				builders.Mesh().
					WithName(zoneProxyMesh).
					WithoutInitialPolicies(),
			),
		).
		Install(MeshTrafficPermissionAllowAllUniversal(zoneProxyMesh)).
		Install(zoneproxy.Install(
			zoneproxy.WithMesh(zoneProxyMesh),
			zoneproxy.WithIngressPort(11001),
			zoneproxy.WithWorkload(zoneProxyIngressDP),
			zoneproxy.WithDpEnvs(zoneProxyDpEnvs),
		)).
		Install(zoneproxy.Install(
			zoneproxy.WithMesh(zoneProxyMesh),
			zoneproxy.WithEgressPort(11002),
			zoneproxy.WithWorkload(zoneProxyEgressDP),
			zoneproxy.WithDpEnvs(zoneProxyDpEnvs),
		)).
		Setup(universal.Cluster)
	Expect(err).ToNot(HaveOccurred())

	// Wait until both zone-proxy DPPs have a usable DataplaneInsight. Without
	// the insight, `kumactl inspect dataplane --shadow` returns 404 and the
	// table tests would only see that error during their (shorter) polling
	// window. Doing the wait here once amortizes the cost over all entries.
	WaitDataplaneInspectable(universal.Cluster, zoneProxyMesh, zoneProxyIngressDP)
	WaitDataplaneInspectable(universal.Cluster, zoneProxyMesh, zoneProxyEgressDP)
}

func CleanupAfterZoneProxyTest(policies ...core_model.ResourceTypeDescriptor) func() {
	return cleanupAfterTest(zoneProxyMesh, []string{zoneProxyIngressDP, zoneProxyEgressDP}, policies...)
}

func CleanupAfterZoneProxySuite() {
	Expect(universal.Cluster.DeleteMeshApps(zoneProxyMesh)).To(Succeed())
	Expect(universal.Cluster.DeleteMesh(zoneProxyMesh)).To(Succeed())
}

func AfterZoneProxyFailure() {
	DebugUniversal(universal.Cluster, zoneProxyMesh)
}
