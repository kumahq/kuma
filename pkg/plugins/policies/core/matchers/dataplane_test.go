package matchers_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	meshexternalservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_matchers "github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_resources "github.com/kumahq/kuma/v3/pkg/test/resources"
	"github.com/kumahq/kuma/v3/pkg/test/resources/builders"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
)

var _ = Describe("MatchedPolicies", func() {
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

var _ = Describe("DppSelectedByPolicy MeshHTTPRoute namespace scoping", func() {
	// route with a shared display-name across namespaces, selecting dataplanes
	// by the given app label.
	route := func(name, namespace, appLabel string) *v1alpha1.MeshHTTPRouteResource {
		return &v1alpha1.MeshHTTPRouteResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh-1",
				Name: name,
				Labels: map[string]string{
					mesh_proto.DisplayName:      "route-1",
					mesh_proto.KubeNamespaceTag: namespace,
				},
			},
			Spec: &v1alpha1.MeshHTTPRoute{
				TargetRef: pointer.To(common_api.TargetRef{
					Kind:   common_api.Dataplane,
					Labels: pointer.To(map[string]string{"app": appLabel}),
				}),
			},
		}
	}

	resources := xds_context.Resources{
		MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
			v1alpha1.MeshHTTPRouteType: &v1alpha1.MeshHTTPRouteResourceList{
				Items: []*v1alpha1.MeshHTTPRouteResource{
					route("route-1.ns-a", "ns-a", "foo"),
					route("route-1.ns-b", "ns-b", "bar"),
				},
			},
		},
	}

	// dpp lives in ns-a but carries app=bar, so it is only reachable through
	// the ns-b route.
	dpp := builders.Dataplane().
		WithName("dp-1").
		WithMesh("mesh-1").
		WithLabels(map[string]string{mesh_proto.KubeNamespaceTag: "ns-a", "app": "bar"}).
		AddInboundOfService("backend").
		Build()

	ref := common_api.TargetRef{
		Kind:   common_api.MeshHTTPRoute,
		Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "route-1"}),
	}

	It("does not leak a namespaced (consumer) policy across namespaces", func() {
		meta := &test_model.ResourceMeta{
			Mesh: "mesh-1",
			Name: "timeout-1",
			Labels: map[string]string{
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.ConsumerPolicyRole),
				mesh_proto.KubeNamespaceTag: "ns-a",
			},
		}

		inbounds, _, err := matchers.DppSelectedByPolicy(meta, ref, dpp, resources)
		Expect(err).ToNot(HaveOccurred())
		Expect(inbounds).To(BeEmpty())
	})

	It("selects through any matching route for a namespace-agnostic (system) policy", func() {
		meta := &test_model.ResourceMeta{Mesh: "mesh-1", Name: "timeout-1"}

		inbounds, _, err := matchers.DppSelectedByPolicy(meta, ref, dpp, resources)
		Expect(err).ToNot(HaveOccurred())
		Expect(inbounds).ToNot(BeEmpty())
	})
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
