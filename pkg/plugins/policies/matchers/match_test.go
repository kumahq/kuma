package matchers_test

import (
	"bytes"
	"fmt"
	"os"
	"path"

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

	readDPP := func(file string) *core_mesh.DataplaneResource {
		dppYaml, err := os.ReadFile(path.Join("testdata", file))
		Expect(err).ToNot(HaveOccurred())

		dpp, err := rest.YAML.UnmarshalCore(dppYaml)
		Expect(err).ToNot(HaveOccurred())
		return dpp.(*core_mesh.DataplaneResource)
	}

	readResourceListResponse := func(file string) xds_context.Resources {
		responseBytes, err := os.ReadFile(path.Join("testdata", file))
		Expect(err).ToNot(HaveOccurred())

		rawResources := yaml.SplitYAML(string(responseBytes))
		resourceList := &policies_api.MeshTrafficPermissionResourceList{}
		for _, rawResource := range rawResources {
			resource, err := rest.YAML.UnmarshalCore([]byte(rawResource))
			Expect(err).ToNot(HaveOccurred())
			err = resourceList.AddItem(resource)
			Expect(err).ToNot(HaveOccurred())
		}

		return xds_context.Resources{
			MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
				policies_api.MeshTrafficPermissionType: resourceList,
			},
		}
	}

	printPolicies := func(policies []core_model.Resource) string {
		matchedPolicyList := &policies_api.MeshTrafficPermissionResourceList{}
		for _, policy := range policies {
			Expect(matchedPolicyList.AddItem(policy)).To(Succeed())
		}
		bytesBuffer := &bytes.Buffer{}
		err := kubectl_output.NewPrinter().Print(rest.From.ResourceList(matchedPolicyList), bytesBuffer)
		Expect(err).ToNot(HaveOccurred())
		return bytesBuffer.String()
	}

	Describe("MatchedDataplanePolicies", func() {

		type testCase struct {
			dppFile                    string
			resourceListResponseFile   string
			matchedResourcesGoldenFile string
		}

		DescribeTable("should return a list of policies ordered by levels for the given DPP",
			func(given testCase) {
				// given DPP resource
				dpp := readDPP(given.dppFile)

				// given MeshTrafficPermissions
				resources := readResourceListResponse(given.resourceListResponseFile)

				// when
				policies, err := matchers.MatchedDataplanePolicies(policies_api.MeshTrafficPermissionType, dpp, resources)
				Expect(err).ToNot(HaveOccurred())

				// then
				Expect(printPolicies(policies.DataplanePolicies)).To(test_matchers.MatchGoldenYAML(path.Join("testdata", given.matchedResourcesGoldenFile)))
			},
			Entry("01. policies select the dataplane without collisions", testCase{
				dppFile:                    "01.dataplane.yaml",
				resourceListResponseFile:   "01.response.yaml",
				matchedResourcesGoldenFile: "01.golden.yaml",
			}),
			Entry("02. policies with the same levels select the dataplane", testCase{
				dppFile:                    "02.dataplane.yaml",
				resourceListResponseFile:   "02.response.yaml",
				matchedResourcesGoldenFile: "02.golden.yaml",
			}),
			Entry("03. policy doesn't selects a union of inbound tags", testCase{
				dppFile:                    "03.dataplane.yaml",
				resourceListResponseFile:   "03.response.yaml",
				matchedResourcesGoldenFile: "03.golden.yaml",
			}),
			Entry("04. policy doesn't select dataplane if at least one tag is not presented", testCase{
				dppFile:                    "04.dataplane.yaml",
				resourceListResponseFile:   "04.response.yaml",
				matchedResourcesGoldenFile: "04.golden.yaml",
			}),
		)
	})

	Describe("MatchedInboundPolicies", func() {

		type testCase struct {
			dppFile                  string
			resourceListResponseFile string
			inboundToGoldenFile      map[string]string
		}

		DescribeTable("should return a map of inbounds with matched policies",
			func(given testCase) {
				// given DPP resource
				dpp := readDPP(given.dppFile)

				// given MeshTrafficPermissions
				resources := readResourceListResponse(given.resourceListResponseFile)

				// when
				policies, err := matchers.MatchedInboundPolicies(policies_api.MeshTrafficPermissionType, dpp, resources)
				Expect(err).ToNot(HaveOccurred())

				// then
				for inbound, matched := range policies.InboundPolicies {
					key := fmt.Sprintf("%v", inbound)
					Expect(printPolicies(matched)).To(test_matchers.MatchGoldenYAML(path.Join("testdata", given.inboundToGoldenFile[key])))
				}
			},
			Entry("05. each inbound is selected by policies", testCase{
				dppFile:                  "05.dataplane.yaml",
				resourceListResponseFile: "05.response.yaml",
				inboundToGoldenFile: map[string]string{
					":8080:8080": "05.golden.8080.yaml",
					":8081:8081": "05.golden.8081.yaml",
				},
			}))
	})

})
