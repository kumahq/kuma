package envoyconfig

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	"github.com/kumahq/kuma/pkg/util/pointer"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
	"slices"
)

func EnvoyConfigTest() {
	meshName := "envoyconfig"

	BeforeAll(func() {
		err := NewClusterSetup().
			Install(
				Yaml(
					builders.Mesh().
						WithName(meshName).
						WithoutInitialPolicies().
						WithMeshServicesEnabled(mesh_proto.Mesh_MeshServices_Exclusive),
				),
			).
			Install(DemoClientUniversal("demo-client", meshName,
				WithTransparentProxy(true)),
			).
			Install(TestServerUniversal("test-server", meshName,
				WithArgs([]string{"echo", "--instance", "universal-1"})),
			).
			Setup(universal.Cluster)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func(g Gomega) {
			_, err := client.CollectEchoResponse(universal.Cluster, "demo-client", "test-server.svc.mesh.local")
			g.Expect(err).ToNot(HaveOccurred())
		}).Should(Succeed())
	})

	AfterEachFailure(func() {
		DebugUniversal(universal.Cluster, meshName)
	})

	E2EAfterAll(func() {
		Expect(universal.Cluster.DeleteMeshApps(meshName)).To(Succeed())
		Expect(universal.Cluster.DeleteMesh(meshName)).To(Succeed())
	})

	E2EAfterEach(func() {
		// delete all meshtimeout policies
		out, err := universal.Cluster.GetKumactlOptions().RunKumactlAndGetOutput("get", "meshtimeouts", "--mesh", meshName, "-o", "json")
		Expect(err).ToNot(HaveOccurred())
		var output struct {
			Items []struct {
				Name string `json:"name"`
			} `json:"items"`
		}
		Expect(json.Unmarshal([]byte(out), &output)).To(Succeed())
		for _, item := range output.Items {
			Expect(universal.Cluster.GetKumactlOptions().RunKumactl("delete", "meshtimeout", item.Name, "--mesh", meshName)).To(Succeed())
		}
	})

	getConfig := func(dpp string) string {
		output, err := universal.Cluster.GetKumactlOptions().
			RunKumactlAndGetOutput("inspect", "dataplane", dpp, "--type", "config", "--mesh", meshName, "--shadow", "--include=diff")
		Expect(err).ToNot(HaveOccurred())
		redacted := redactIPs(output)

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
			Expect(getConfig("demo-client")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "demo-client.golden.json", 1)))
			Expect(getConfig("test-server")).To(matchers.MatchGoldenJSON(strings.Replace(inputFile, "input.yaml", "test-server.golden.json", 1)))
		}, test.EntriesForFolder("meshtimeout", "envoyconfig"),
	)
}

var ipv6Regex = `\[?` + // Optional opening square bracket for IPv6 in URLs (e.g., [2001:db8::1])
	`(` +
	// Full IPv6 address with 8 segments (e.g., 2001:0db8:85a3:0000:0000:8a2e:0370:7334)
	`([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|` +
	// IPv6 with leading compression (e.g., ::1, ::8a2e:0370:7334)
	`([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|` +
	// IPv6 with trailing compression (e.g., 2001:db8::, 2001:db8::1:2)
	`([0-9a-fA-F]{1,4}:){1,7}:|` +
	// IPv6 with mixed compression (e.g., 2001:db8:0:0::1:2)
	`([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|` +
	`([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|` +
	`([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|` +
	`([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|` +
	// IPv6 with only one segment and compression (e.g., 2001::1:2:3:4:5:6)
	`[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|` +
	// Fully compressed IPv6 (::) or with trailing segments (::1, ::8a2e:0370:7334)
	`:((:[0-9a-fA-F]{1,4}){1,7}|:)|` +
	// Link-local IPv6 with zone identifiers (e.g., fe80::1%eth0)
	`fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]+|` +
	// IPv6 with embedded IPv4 (e.g., ::ffff:192.168.1.1)
	`::(ffff(:0{1,4})?:)?((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}` +
	`(25[0-5]|(2[0-4]|1?[0-9])?[0-9])|` +
	// Mixed IPv6 and IPv4 (e.g., 2001:db8::192.168.1.1)
	`([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1?[0-9])?[0-9])\.){3}` +
	`(25[0-5]|(2[0-4]|1?[0-9])?[0-9])` +
	`)` +
	`]?` // Optional closing square bracket for IPv6 in URLs (e.g., [2001:db8::1])

var ipv4Regex = `\b` + // Word boundary to ensure we match standalone IPv4 addresses
	`\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}` + // Matches IPv4 format (e.g., 192.168.0.1)
	`\b`

var ipRegex = regexp.MustCompile(ipv4Regex + "|" + ipv6Regex)

func redactIPs(jsonStr string) string {
	return ipRegex.ReplaceAllString(jsonStr, "IP_REDACTED")
}
