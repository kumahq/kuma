package envoyconfig

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshaccesslog "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshcircuitbreaker "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	meshfaultinjection "github.com/kumahq/kuma/pkg/plugins/policies/meshfaultinjection/api/v1alpha1"
	meshhealthcheck "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	meshhttproute "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshratelimit "github.com/kumahq/kuma/pkg/plugins/policies/meshratelimit/api/v1alpha1"
	meshretry "github.com/kumahq/kuma/pkg/plugins/policies/meshretry/api/v1alpha1"
	meshtimeout "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtls "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

const meshName = "envoyconfig"

func Sidecars() {
	BeforeAll(SetupSidecarCluster)

	AfterEachFailure(AfterSidecarFailure)

	E2EAfterAll(CleanupAfterSidecarSuite)

	E2EAfterEach(CleanupAfterSidecarTest(
		meshtimeout.MeshTimeoutResourceTypeDescriptor,
		meshaccesslog.MeshAccessLogResourceTypeDescriptor,
		meshfaultinjection.MeshFaultInjectionResourceTypeDescriptor,
		meshratelimit.MeshRateLimitResourceTypeDescriptor,
		meshtls.MeshTLSResourceTypeDescriptor,
		meshhealthcheck.MeshHealthCheckResourceTypeDescriptor,
		meshcircuitbreaker.MeshCircuitBreakerResourceTypeDescriptor,
		meshhttproute.MeshHTTPRouteResourceTypeDescriptor,
		meshretry.MeshRetryResourceTypeDescriptor,
	))

	DescribeTable("should generate proper Envoy config",
		TestSidecarConfig,
		test.EntriesForFolder(filepath.Join("sidecars", "meshtimeout"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("sidecars", "meshaccesslog"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("sidecars", "meshfaultinjection"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("sidecars", "meshratelimit"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("sidecars", "meshtls"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("sidecars", "meshcircuitbreaker"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("sidecars", "meshretry"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("sidecars", "meshloadbalancingstrategy"), "envoyconfig"),
	)
}

func TestSidecarConfig(inputFile string) {
	// given
	input, err := os.ReadFile(inputFile)
	Expect(err).ToNot(HaveOccurred())

	// when
	if len(input) > 0 {
		Expect(universal.Cluster.Install(YamlUniversal(string(input)))).To(Succeed())
	}

	// then
	Eventually(func(g Gomega) {
		g.Expect(getConfig(meshName, "demo-client")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "demo-client.golden.json", 1)))
		g.Expect(getConfig(meshName, "test-server")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "test-server.golden.json", 1)))
	}).Should(Succeed())
}

func SetupSidecarCluster() {
	err := NewClusterSetup().
		Install(
			Yaml(
				builders.Mesh().
					WithName(meshName).
					WithoutInitialPolicies().
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).
					WithBuiltinMTLSBackend("ca-1").WithEnabledMTLSBackend("ca-1"),
			),
		).
		Install(MeshTrafficPermissionAllowAllUniversal(meshName)).
		Install(DemoClientUniversal("demo-client", meshName,
			WithTransparentProxy(true),
			WithDpEnvs(map[string]string{
				"KUMA_DATAPLANE_RUNTIME_SOCKET_DIR": "/tmp",
			})),
		).
		Install(TestServerUniversal("test-server", meshName,
			WithArgs([]string{"echo", "--instance", "universal-1"}),
			WithDpEnvs(map[string]string{
				"KUMA_DATAPLANE_RUNTIME_SOCKET_DIR": "/tmp",
			}),
		),
		).
		Setup(universal.Cluster)
	Expect(err).ToNot(HaveOccurred())

	waitMeshServiceReady(meshName, "demo-client")
	waitMeshServiceReady(meshName, "test-server")

	Eventually(func(g Gomega) {
		_, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.svc.mesh.local")
		g.Expect(err).ToNot(HaveOccurred())
	}).Should(Succeed())
}

func CleanupAfterSidecarTest(policies ...core_model.ResourceTypeDescriptor) func() {
	return cleanupAfterTest(meshName, policies...)
}

func CleanupAfterSidecarSuite() {
	Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
	Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
}

func AfterSidecarFailure() {
	DebugUniversal(universal.Cluster, meshName)
}
