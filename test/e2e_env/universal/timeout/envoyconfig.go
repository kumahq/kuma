package timeout

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/builders"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/client"
	"github.com/kumahq/kuma/test/framework/envs/universal"
)

func EnvoyConfigTest() {
	meshName := "meshtimeout-envoyconfig"

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

	redactIPs := func(jsonStr string) string {
		ipRegex := regexp.MustCompile(`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`)
		return ipRegex.ReplaceAllString(jsonStr, "IP_REDACTED")
	}

	getConfig := func(dpp string) string {
		output, err := universal.Cluster.GetKumactlOptions().
			RunKumactlAndGetOutput("inspect", "dataplane", dpp, "--type", "config", "--mesh", meshName, "--shadow", "--include=diff")
		Expect(err).ToNot(HaveOccurred())
		return redactIPs(output)
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
		}, test.EntriesForFolder("meshtimeout", "timeout"),
	)
}
