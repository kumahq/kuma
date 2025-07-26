package envoyconfig

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshaccesslog "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtimeout "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	meshtls "github.com/kumahq/kuma/pkg/plugins/policies/meshtls/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

const mesh = "envoyconfig-builtingateway"

const gatewayPort = 8080

var gatewayAddressPort = func(appName string, port int) string {
	ip := universal.Cluster.GetApp(appName).GetIP()
	return net.JoinHostPort(ip, strconv.Itoa(port))
}

var meshGateway = func() string {
	return fmt.Sprintf(`
type: MeshGateway
name: gateway-proxy
mesh: %s
selectors:
- match:
    kuma.io/service: gateway-proxy
conf:
  listeners:
  - port: %d
    protocol: HTTP
    hostname: example.kuma.io
`, mesh, gatewayPort)
}

var meshHTTPRoute = func() string {
	return fmt.Sprintf(`
type: MeshHTTPRoute
name: gateway-proxy-route
mesh: %s
spec:
  targetRef:
    kind: MeshGateway
    name: gateway-proxy
  to:
    - targetRef:
        kind: Mesh
      rules:
        - matches:
            - path:
                type: PathPrefix
                value: "/"
          default:
            backendRefs:
              - kind: MeshService
                name: echo-service
                port: 80
`, mesh)
}

func BuiltinGateway() {
	BeforeAll(SetupGatewayCluster)

	AfterEachFailure(AfterGatewayFailure)

	E2EAfterAll(CleanupAfterGatewaySuite)

	E2EAfterEach(CleanupAfterGatewayTest(
		meshaccesslog.MeshAccessLogResourceTypeDescriptor,
		meshtls.MeshTLSResourceTypeDescriptor,
		meshtimeout.MeshTimeoutResourceTypeDescriptor,
	))

	DescribeTable("should generate proper Envoy config",
		TestBuiltinGatewayConfig,
		test.EntriesForFolder(filepath.Join("builtingateway", "meshaccesslog"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("builtingateway", "meshtls"), "envoyconfig"),
		test.EntriesForFolder(filepath.Join("builtingateway", "meshtimeout"), "envoyconfig"),
	)
}

func TestBuiltinGatewayConfig(inputFile string) {
	// given
	input, err := os.ReadFile(inputFile)
	Expect(err).ToNot(HaveOccurred())

	// when
	if len(input) > 0 {
		Expect(universal.Cluster.Install(YamlUniversal(string(input)))).To(Succeed())
	}

	// then
	Eventually(func(g Gomega) {
		g.Expect(getConfig(mesh, "gateway-proxy")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "gateway-proxy.golden.json", 1)))
	}).Should(Succeed())
}

func SetupGatewayCluster() {
	setup := NewClusterSetup().
		Install(
			Yaml(
				builders.Mesh().
					WithName(mesh).
					WithoutInitialPolicies().
					WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive).
					WithBuiltinMTLSBackend("ca-1").WithEnabledMTLSBackend("ca-1"),
			),
		).
		Install(MeshTrafficPermissionAllowAllUniversal(mesh)).
		Install(GatewayClientAppUniversal("gateway-client")).
		Install(EchoServerApp(mesh, "echo-server", "echo-service", "universal")).
		Install(GatewayProxyUniversal(mesh, "gateway-proxy", WithDpEnvs(map[string]string{
			"KUMA_DATAPLANE_RUNTIME_SOCKET_DIR": "/tmp",
		}))).
		Install(YamlUniversal(meshGateway())).
		Install(YamlUniversal(meshHTTPRoute()))

	Expect(setup.Setup(universal.Cluster)).To(Succeed())

	waitMeshServiceReady(mesh, "echo-service")

	Eventually(ProxySimpleRequests(universal.Cluster, "universal",
		gatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io"), "60s", "1s").Should(Succeed())
}

func CleanupAfterGatewayTest(policies ...core_model.ResourceTypeDescriptor) func() {
	return cleanupAfterTest(mesh, policies...)
}

func CleanupAfterGatewaySuite() {
	Expect(universal.Cluster.DeleteApp("gateway-client")).To(Succeed())
	Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
	Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
}

func AfterGatewayFailure() {
	DebugUniversal(universal.Cluster, mesh)
}
