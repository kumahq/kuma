package matchers_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	kubectl_output "github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	_ "github.com/kumahq/kuma/pkg/plugins/policies"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("Match", func() {
	readDPP := func(file string) *core_mesh.DataplaneResource {
		dppYaml, err := os.ReadFile(file)
		Expect(err).ToNot(HaveOccurred())

		dpp, err := rest.YAML.UnmarshalCore(dppYaml)
		Expect(err).ToNot(HaveOccurred())
		return dpp.(*core_mesh.DataplaneResource)
	}

	readPolicies := func(file string) xds_context.Resources {
		responseBytes, err := os.ReadFile(file)
		Expect(err).ToNot(HaveOccurred())

		rawResources := util_yaml.SplitYAML(string(responseBytes))

		var trafficPermissions []*policies_api.MeshTrafficPermissionResource
		var gateways []*core_mesh.MeshGatewayResource

		for _, rawResource := range rawResources {
			resource, err := rest.YAML.UnmarshalCore([]byte(rawResource))
			Expect(err).ToNot(HaveOccurred())

			switch resource.Descriptor().Name {
			case policies_api.MeshTrafficPermissionType:
				trafficPermissions = append(trafficPermissions, resource.(*policies_api.MeshTrafficPermissionResource))
			case core_mesh.MeshGatewayType:
				gateways = append(gateways, resource.(*core_mesh.MeshGatewayResource))
			}
		}

		return xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				policies_api.MeshTrafficPermissionType: &policies_api.MeshTrafficPermissionResourceList{
					Items: trafficPermissions,
				},
				core_mesh.MeshGatewayType: &core_mesh.MeshGatewayResourceList{
					Items: gateways,
				},
			},
		}
	}

	Describe("MatchedPolicies", func() {
		type testCase struct {
			dppFile      string
			policiesFile string
			goldenFile   string
		}

		generateTableEntries := func(testDir string) []TableEntry {
			var res []TableEntry
			files, err := os.ReadDir(testDir)
			Expect(err).ToNot(HaveOccurred())

			testCaseMap := map[string]*testCase{}
			for _, f := range files {
				parts := strings.Split(f.Name(), ".")
				// file name has a format 01.golden.yaml
				num, fileType := parts[0], parts[1]
				if _, ok := testCaseMap[num]; !ok {
					testCaseMap[num] = &testCase{}
				}
				switch fileType {
				case "dataplane":
					testCaseMap[num].dppFile = filepath.Join(testDir, f.Name())
				case "policies":
					testCaseMap[num].policiesFile = filepath.Join(testDir, f.Name())
				case "golden":
					testCaseMap[num].goldenFile = filepath.Join(testDir, f.Name())
				}
			}

			for num, tc := range testCaseMap {
				res = append(res, Entry(num, *tc))
			}
			return res
		}

		DescribeTable("should return a list of DataplanePolicies ordered by levels for the given DPP",
			func(given testCase) {
				// given DPP resource
				dpp := readDPP(given.dppFile)

				// given MeshTrafficPermissions
				resources := readPolicies(given.policiesFile)

				// when
				policies, err := matchers.MatchedPolicies(policies_api.MeshTrafficPermissionType, dpp, resources)
				Expect(err).ToNot(HaveOccurred())

				// then
				matchedPolicyList := &policies_api.MeshTrafficPermissionResourceList{}
				for _, policy := range policies.DataplanePolicies {
					Expect(matchedPolicyList.AddItem(policy)).To(Succeed())
				}
				bytesBuffer := &bytes.Buffer{}
				err = kubectl_output.NewPrinter().Print(rest.From.ResourceList(matchedPolicyList), bytesBuffer)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytesBuffer.String()).To(test_matchers.MatchGoldenYAML(given.goldenFile))
			},
			generateTableEntries(filepath.Join("testdata", "match", "dataplanepolicies")),
		)

		DescribeTable("should return FromRules",
			func(given testCase) {
				// given DPP resource
				dpp := readDPP(given.dppFile)

				// given MeshTrafficPermissions
				resources := readPolicies(given.policiesFile)

				// when
				policies, err := matchers.MatchedPolicies(policies_api.MeshTrafficPermissionType, dpp, resources)
				Expect(err).ToNot(HaveOccurred())

				// then
				bytes, err := yaml.Marshal(policies.FromRules)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
			},
			generateTableEntries(filepath.Join("testdata", "match", "fromrules")),
		)

		DescribeTable("should match MeshGateways",
			func(given testCase) {
				dpp := readDPP(given.dppFile)

				resources := readPolicies(given.policiesFile)

				policies, err := matchers.MatchedPolicies(policies_api.MeshTrafficPermissionType, dpp, resources)
				Expect(err).ToNot(HaveOccurred())

				bytes, err := yaml.Marshal(policies.FromRules)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
			},
			generateTableEntries(filepath.Join("testdata", "match", "meshgateways")),
		)
	})
})
