package matchers_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kubectl_output "github.com/kumahq/kuma/app/kumactl/pkg/output/yaml"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	_ "github.com/kumahq/kuma/pkg/plugins/policies"
	"github.com/kumahq/kuma/pkg/plugins/policies/matchers"
	policies_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("Match", func() {

	Describe("MatchedPolicies", func() {

		type testCase struct {
			dppFile                    string
			resourceListResponseFile   string
			matchedResourcesGoldenFile string
		}

		DescribeTable("should return a list of policies ordered by levels for the given DPP",
			func(given testCase) {
				// given DPP resource
				dppYaml, err := os.ReadFile(given.dppFile)
				Expect(err).ToNot(HaveOccurred())

				resCore, err := rest.YAML.UnmarshalCore(dppYaml)
				Expect(err).ToNot(HaveOccurred())
				dpp := resCore.(*core_mesh.DataplaneResource)

				// given MeshTrafficPermissions
				responseBytes, err := os.ReadFile(given.resourceListResponseFile)
				Expect(err).ToNot(HaveOccurred())

				rawResources := yaml.SplitYAML(string(responseBytes))
				resourceList := &policies_api.MeshTrafficPermissionResourceList{}
				for _, rawResource := range rawResources {
					resource, err := rest.YAML.UnmarshalCore([]byte(rawResource))
					Expect(err).ToNot(HaveOccurred())
					err = resourceList.AddItem(resource)
					Expect(err).ToNot(HaveOccurred())
				}

				resources := xds_context.Resources{
					MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
						policies_api.MeshTrafficPermissionType: resourceList,
					},
				}

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
				Expect(bytesBuffer.String()).To(test_matchers.MatchGoldenYAML(given.matchedResourcesGoldenFile))
			}, func() []TableEntry {
				var res []TableEntry
				testDir := filepath.Join("testdata", "match")
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
						testCaseMap[num].resourceListResponseFile = filepath.Join(testDir, f.Name())
					case "golden":
						testCaseMap[num].matchedResourcesGoldenFile = filepath.Join(testDir, f.Name())
					}
				}

				for num, tc := range testCaseMap {
					res = append(res, Entry(num, *tc))
				}
				return res
			}(),
		)
	})
})
