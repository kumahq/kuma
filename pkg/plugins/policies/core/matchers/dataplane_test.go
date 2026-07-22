package matchers_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtls_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtls/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_matchers "github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_resources "github.com/kumahq/kuma/v3/pkg/test/resources"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ = Describe("MatchedPolicies", func() {
	It("returns GatewayRules for mesh-wide policy on a legacy built-in gateway with no inbounds", func() {
		// given
		dpp := &core_mesh.DataplaneResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh-1",
				Name: "builtin-gateway",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.0.2.1",
					Gateway: &mesh_proto.Dataplane_Networking_Gateway{
						Type: mesh_proto.Dataplane_Networking_Gateway_BUILTIN,
						Tags: map[string]string{
							mesh_proto.ServiceTag: "gateway",
						},
					},
				},
			},
		}

		mode := meshtls_api.ModeStrict
		policy := meshtls_api.NewMeshTLSResource()
		policy.Meta = &test_model.ResourceMeta{
			Mesh: "mesh-1",
			Name: "mesh-wide",
		}
		policy.Spec.TargetRef = &common_api.TargetRef{Kind: common_api.Mesh}
		policy.Spec.Rules = &[]meshtls_api.Rule{{
			Default: meshtls_api.Conf{
				Mode: &mode,
			},
		}}
		resources := xds_context.NewResources()
		resources.MeshLocalResources[meshtls_api.MeshTLSType] = &meshtls_api.MeshTLSResourceList{
			Items: []*meshtls_api.MeshTLSResource{policy},
		}

		// when
		matches, err := matchers.PolicyMatches(policy, dpp, resources)
		Expect(err).ToNot(HaveOccurred())
		policies, err := matchers.MatchedPolicies(meshtls_api.MeshTLSType, dpp, resources)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(matches).To(BeTrue())
		Expect(policies.DataplanePolicies).To(ConsistOf(policy))
		Expect(policies.FromRules.InboundRules).To(BeEmpty())
		Expect(policies.GatewayRules.InboundRules).To(HaveKey(core_rules.InboundListener{
			Address: "192.0.2.1",
		}))
	})

	type testCase struct {
		testName     string
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
			if len(parts) < 2 {
				continue
			}
			// file name has a format 01.golden.yaml
			name, fileType := parts[0], parts[1]
			if _, ok := testCaseMap[name]; !ok {
				testCaseMap[name] = &testCase{}
				testCaseMap[name].testName = name
			}
			switch fileType {
			case "dataplane":
				testCaseMap[name].dppFile = filepath.Join(testDir, f.Name())
			case "policies":
				testCaseMap[name].policiesFile = filepath.Join(testDir, f.Name())
			case "golden":
				testCaseMap[name].goldenFile = filepath.Join(testDir, f.Name())
			case "mes":
				testCaseMap[name].mesFile = filepath.Join(testDir, f.Name())
			}
		}

		for _, tc := range testCaseMap {
			res = append(res, Entry(tc.testName, *tc))
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

			resType := getResourceType(resTypes)

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

	type dataplaneTestCase struct {
		dataplaneMeta test_resources.BuildMeta
		policyMeta    test_resources.BuildMeta
		goldenFile    string
	}
	DescribeTableSubtree("should match by kind Dataplane", func(givenResources testCase) {
		DescribeTable("should TODO", func(given dataplaneTestCase) {
			// given
			dpp := readDPP(givenResources.dppFile)
			test_resources.UpdateResourceMeta(given.dataplaneMeta, dpp)

			resources, resTypes := readPolicies(givenResources.policiesFile)

			resType := getResourceType(resTypes)
			test_resources.UpdateResourcesMeta(given.policyMeta, resources.MeshLocalResources[resType])

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
			Entry("uni zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneUni,
				policyMeta:    test_resources.ZoneUni,
				goldenFile:    buildGoldenFilePath("uni-zone", givenResources.testName),
			}),
			Entry("k8s zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneK8s,
				policyMeta:    test_resources.ZoneK8s,
				goldenFile:    buildGoldenFilePath("k8s-zone", givenResources.testName),
			}),
			Entry("policy global uni, dpp uni - on global", dataplaneTestCase{
				dataplaneMeta: test_resources.SyncToUni(test_resources.ZoneUni),
				policyMeta:    test_resources.SystemPolicy(test_resources.GlobalUni),
				goldenFile:    buildGoldenFilePath("policy-from-global-uni-zone-uni-on-global", givenResources.testName),
			}),
			Entry("policy global uni, dpp uni - on zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneUni,
				policyMeta:    test_resources.SystemPolicy(test_resources.SyncToUni(test_resources.GlobalUni)),
				goldenFile:    buildGoldenFilePath("policy-from-global-uni-zone-uni-on-zone", givenResources.testName),
			}),
			Entry("policy global uni, dpp k8s - on zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneK8s,
				policyMeta:    test_resources.SystemPolicy(test_resources.SyncToK8s(test_resources.GlobalUni)),
				goldenFile:    buildGoldenFilePath("policy-from-global-uni-zone-k8s-on-zone", givenResources.testName),
			}),
			Entry("policy global uni, dpp k8s - on global", dataplaneTestCase{
				dataplaneMeta: test_resources.SyncToUni(test_resources.ZoneK8s),
				policyMeta:    test_resources.SystemPolicy(test_resources.GlobalUni),
				goldenFile:    buildGoldenFilePath("policy-from-global-uni-zone-k8s-on-global", givenResources.testName),
			}),
			Entry("policy global k8s, dpp uni - on zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneUni,
				policyMeta:    test_resources.SystemPolicy(test_resources.SyncToUni(test_resources.GlobalK8s)),
				goldenFile:    buildGoldenFilePath("policy-from-global-k8s-zone-uni-on-zone", givenResources.testName),
			}),
			Entry("policy global k8s, dpp uni - on global", dataplaneTestCase{
				dataplaneMeta: test_resources.SyncToK8s(test_resources.ZoneUni),
				policyMeta:    test_resources.SystemPolicy(test_resources.GlobalK8s),
				goldenFile:    buildGoldenFilePath("policy-from-global-k8s-zone-uni-on-global", givenResources.testName),
			}),
			Entry("policy global k8s, dpp k8s - on zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneK8s,
				policyMeta:    test_resources.SystemPolicy(test_resources.SyncToK8s(test_resources.GlobalK8s)),
				goldenFile:    buildGoldenFilePath("policy-from-global-k8s-zone-k8s-on-zone", givenResources.testName),
			}),
			Entry("policy global k8s, dpp k8s - on global", dataplaneTestCase{
				dataplaneMeta: test_resources.SyncToK8s(test_resources.ZoneK8s),
				policyMeta:    test_resources.SystemPolicy(test_resources.GlobalK8s),
				goldenFile:    buildGoldenFilePath("policy-from-global-k8s-zone-k8s-on-global", givenResources.testName),
			}),
			Entry("policy global k8s, dpp uni - on zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneUni,
				policyMeta:    test_resources.SystemPolicy(test_resources.SyncToUni(test_resources.GlobalUni)),
				goldenFile:    buildGoldenFilePath("policy-global-uni-dpp-k8s-on-zone", givenResources.testName),
			}),
			Entry("policy global k8s, dpp uni - on global", dataplaneTestCase{
				dataplaneMeta: test_resources.SyncToUni(test_resources.ZoneUni),
				policyMeta:    test_resources.SystemPolicy(test_resources.GlobalUni),
				goldenFile:    buildGoldenFilePath("policy-global-uni-dpp-k8s-on-global", givenResources.testName),
			}),
			Entry("policy synced from other k8s zone", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneUni,
				policyMeta:    test_resources.ProducerPolicy(test_resources.SyncToUni(test_resources.ZoneK8s)),
				goldenFile:    buildGoldenFilePath("policy-from-k8s-to-uni", givenResources.testName),
			}),
			Entry("policy synced from other k8s zone to k8s", dataplaneTestCase{
				dataplaneMeta: test_resources.ZoneK8s,
				policyMeta:    test_resources.ProducerPolicy(test_resources.SyncToK8s(test_resources.ZoneK8s)),
				goldenFile:    buildGoldenFilePath("policy-from-k8s-to-k8s", givenResources.testName),
			}),
		)
	}, generateTableEntries(filepath.Join("testdata", "matchedpolicies", "dataplane-kind")))
})

func getResourceType(resTypes []core_model.ResourceType) core_model.ResourceType {
	var resType core_model.ResourceType
	switch {
	case len(resTypes) == 1:
		resType = resTypes[0]
	case len(resTypes) == 2 && resTypes[1] == v1alpha1.MeshHTTPRouteType:
		resType = resTypes[0]
	case len(resTypes) == 2 && resTypes[0] == v1alpha1.MeshHTTPRouteType:
		resType = resTypes[1]
	}
	return resType
}

func buildGoldenFilePath(caseName, testName string) string {
	return filepath.Join("testdata", "matchedpolicies", "dataplane-kind", testName, fmt.Sprintf("%s.golden.yaml", caseName))
}
