package matchers_test

import (
	"bytes"
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
	matchers2 "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/util/yaml"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("MatchedPolicies", func() {

	type testCase struct {
		dppFile                    string
		resourceListResponseFile   string
		matchedResourcesGoldenFile string
	}

	DescribeTable("should return a list of policies ordered by levels for the given DPP",
		func(given testCase) {
			// given DPP resource
			dppYaml, err := os.ReadFile(path.Join("testdata", given.dppFile))
			Expect(err).ToNot(HaveOccurred())

			dpp, err := rest.YAML.UnmarshalCore(dppYaml)
			Expect(err).ToNot(HaveOccurred())

			// given MeshTrafficPermissions
			responseBytes, err := os.ReadFile(path.Join("testdata", given.resourceListResponseFile))
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
			policies, err := matchers.MatchedPolicies(policies_api.MeshTrafficPermissionType, dpp.(*core_mesh.DataplaneResource), resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			matchedPolicyList := &policies_api.MeshTrafficPermissionResourceList{}
			for _, policy := range policies.DataplanePolicies {
				Expect(matchedPolicyList.AddItem(policy)).To(Succeed())
			}
			bytesBuffer := &bytes.Buffer{}
			err = kubectl_output.NewPrinter().Print(rest.From.ResourceList(matchedPolicyList), bytesBuffer)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytesBuffer.String()).To(matchers2.MatchGoldenYAML(path.Join("testdata", given.matchedResourcesGoldenFile)))
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
