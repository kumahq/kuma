package matchers_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	meshaccesslog_api "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("MatchedPolicies", func() {
	type testCase struct {
		dppFile      string
		mesFile      string
		policiesFile string
		goldenFile   string
	}

	generateTableEntries := func(testDir string) []TableEntry {
		defer GinkgoRecover()
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
			case "mes":
				testCaseMap[num].mesFile = filepath.Join(testDir, f.Name())
			}
		}

		for _, tc := range testCaseMap {
			res = append(res, Entry(tc.goldenFile, *tc))
		}
		return res
	}

	DescribeTable("should return a list of DataplanePolicies ordered by levels for the given DPP",
		func(given testCase) {
			// given DPP resource
			dpp := readDPP(given.dppFile)

			// given policies
			resources, resTypes := readPolicies(given.policiesFile)

			// we're expecting all policies in the file to have the same type or to be mixed with MeshHTTPRoutes
			Expect(resTypes).To(Or(HaveLen(1), HaveLen(2)))

			var resType core_model.ResourceType
			switch {
			case len(resTypes) == 1:
				resType = resTypes[0]
			case len(resTypes) == 2 && resTypes[1] == v1alpha1.MeshHTTPRouteType:
				resType = resTypes[0]
			case len(resTypes) == 2 && resTypes[0] == v1alpha1.MeshHTTPRouteType:
				resType = resTypes[1]
			}

			// when
			policies, err := matchers.MatchedPolicies(resType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			matchedPolicyList, err := registry.Global().NewList(resType)
			Expect(err).ToNot(HaveOccurred())

			for _, policy := range policies.DataplanePolicies {
				Expect(matchedPolicyList.AddItem(policy)).To(Succeed())
			}
			bytes, err := yaml.Marshal(rest.From.ResourceList(matchedPolicyList))
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		},
		generateTableEntries(filepath.Join("testdata", "matchedpolicies", "dataplanepolicies")),
	)

	DescribeTable("should return FromRules",
		func(given testCase) {
			// given DPP resource
			dpp := readDPP(given.dppFile)

			// given MeshTrafficPermissions
			resources, _ := readPolicies(given.policiesFile)

			// when
			policies, err := matchers.MatchedPolicies(meshtrafficpermission_api.MeshTrafficPermissionType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(policies.FromRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		},
		generateTableEntries(filepath.Join("testdata", "matchedpolicies", "fromrules")),
	)

	DescribeTable("should return ToRules",
		func(given testCase) {
			// given DPP resource
			dpp := readDPP(given.dppFile)

			// given policies
			resources, resTypes := readPolicies(given.policiesFile)

			// we're expecting all policies in the file to have the same type or to be mixed with MeshHTTPRoutes
			Expect(resTypes).To(Or(HaveLen(1), HaveLen(2)))

			var resType core_model.ResourceType
			switch {
			case len(resTypes) == 1:
				resType = resTypes[0]
			case len(resTypes) == 2 && resTypes[1] == v1alpha1.MeshHTTPRouteType:
				resType = resTypes[0]
			case len(resTypes) == 2 && resTypes[0] == v1alpha1.MeshHTTPRouteType:
				resType = resTypes[1]
			}

			// when
			policies, err := matchers.MatchedPolicies(resType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(policies.ToRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		},
		generateTableEntries(filepath.Join("testdata", "matchedpolicies", "torules")),
	)

	DescribeTable("should return ToRules for MeshExternalService",
		func(given testCase) {
			// given DPP resource
			dpp := readDPP(given.dppFile)

			// given policies
			resources, resTypes := readPolicies(given.policiesFile)

			// given MeshExternalService resource
			mes := readMES(given.mesFile)
			resources.MeshLocalResources[meshexternalservice_api.MeshExternalServiceType] = &meshexternalservice_api.MeshExternalServiceResourceList{
				Items: []*meshexternalservice_api.MeshExternalServiceResource{mes},
			}

			// we're expecting all policies in the file to have the same type or to be mixed with MeshHTTPRoutes
			Expect(resTypes).To(Or(HaveLen(1), HaveLen(2)))

			var resType core_model.ResourceType
			switch {
			case len(resTypes) == 1:
				resType = resTypes[0]
			case len(resTypes) == 2 && resTypes[1] == v1alpha1.MeshHTTPRouteType:
				resType = resTypes[0]
			case len(resTypes) == 2 && resTypes[0] == v1alpha1.MeshHTTPRouteType:
				resType = resTypes[1]
			}

			// when
			policies, err := matchers.MatchedPolicies(resType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(policies.ToRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		},
		generateTableEntries(filepath.Join("testdata", "matchedpolicies", "meshexternalservice")),
	)

	DescribeTable("should match MeshGateways",
		func(given testCase) {
			dpp := readDPP(given.dppFile)

			resources, _ := readPolicies(given.policiesFile)

			policies, err := matchers.MatchedPolicies(meshaccesslog_api.MeshAccessLogType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())

			bytes, err := yaml.Marshal(policies.GatewayRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		},
		generateTableEntries(filepath.Join("testdata", "matchedpolicies", "meshgateways")),
	)

	DescribeTable("should match MeshGateways for dpp without meta",
		func(given testCase) {
			dpp := readDPP(given.dppFile)
			dpp.SetMeta(nil)

			resources, _ := readPolicies(given.policiesFile)

			policies, err := matchers.MatchedPolicies(meshaccesslog_api.MeshAccessLogType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())

			bytes, err := yaml.Marshal(policies.GatewayRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		},
		generateTableEntries(filepath.Join("testdata", "matchedpolicies", "dpp_without_meta")),
	)
})
