package envoyconfig

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	meshtimeout "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/e2e_env/universal/gateway"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func BuiltinGateway() {
	mesh := "envoyconfig-builtingateway"

	const gatewayPort = 8080

	gatewayAddressPort := func(appName string, port int) string {
		ip := universal.Cluster.GetApp(appName).GetIP()
		return net.JoinHostPort(ip, strconv.Itoa(port))
	}

	meshGateway := func() string {
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

	meshHTTPRoute := func() string {
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

	BeforeAll(func() {
		setup := NewClusterSetup().
			Install(
				Yaml(
					builders.Mesh().
						WithName(mesh).
						WithoutInitialPolicies().
						WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
				),
			).
			Install(GatewayClientAppUniversal("gateway-client")).
			Install(EchoServerApp(mesh, "echo-server", "echo-service", "universal")).
			Install(GatewayProxyUniversal(mesh, "gateway-proxy")).
			Install(YamlUniversal(meshGateway())).
			Install(YamlUniversal(meshHTTPRoute()))

		Expect(setup.Setup(universal.Cluster)).To(Succeed())

		Eventually(ProxySimpleRequests(universal.Cluster, "universal",
			gatewayAddressPort("gateway-proxy", gatewayPort), "example.kuma.io"), "60s", "1s").Should(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, mesh)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteApp("gateway-client")).To(Succeed())
		Expect(universal.Cluster.DeleteMeshApps(mesh)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(mesh)).To(Succeed())
	})

	E2EAfterEach(func() {
		Expect(DeleteMeshResources(
			universal.Cluster,
			mesh,
			meshtimeout.MeshTimeoutResourceTypeDescriptor,
		)).To(Succeed())
	})

	getConfig := func(dpp string) string {
		output, err := universal.Cluster.GetKumactlOptions().
			RunKumactlAndGetOutput("inspect", "dataplane", dpp, "--type", "config", "--mesh", mesh, "--shadow", "--include=diff")
		Expect(err).ToNot(HaveOccurred())
		redacted := redactStatPrefixes(redactIPs(output))

		response := types.InspectDataplanesConfig{}
		Expect(json.Unmarshal([]byte(redacted), &response)).To(Succeed())
		Expect(response.Diff).ToNot(BeNil())
		response.Diff = pointer.To(slices.DeleteFunc(*response.Diff, func(item api_common.JsonPatchItem) bool {
			return item.Op == api_common.Test
		}))
		slices.SortStableFunc(*response.Diff, func(a, b api_common.JsonPatchItem) int {
			return strings.Compare(a.Path, b.Path)
		})

		result, err := json.MarshalIndent(response, "", "  ")
		Expect(err).ToNot(HaveOccurred())
		return string(result)
	}

	DescribeTable("should generate proper Envoy config",
		func(inputFile string) {
			// given
			input, err := os.ReadFile(inputFile)
			Expect(err).ToNot(HaveOccurred())

			// when
			if len(input) > 0 {
				Expect(universal.Cluster.Install(YamlUniversal(string(input)))).To(Succeed())
			}

			// then
			Expect(getConfig("gateway-proxy")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "gateway-proxy.golden.json", 1)))
		},
		test.EntriesForFolder(filepath.Join("builtingateway", "meshtimeout"), "envoyconfig"),
	)
}
