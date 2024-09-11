package matchers_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	mcb_api "github.com/kumahq/kuma/pkg/plugins/policies/meshcircuitbreaker/api/v1alpha1"
	mhc_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhealthcheck/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	mt_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/file"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
)

var _ = Describe("EgressMatchedPolicies", func() {
	type testCase struct {
		esFile       string
		mesFile      string
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
			case "es":
				testCaseMap[num].esFile = filepath.Join(testDir, f.Name())
			case "mes":
				testCaseMap[num].mesFile = filepath.Join(testDir, f.Name())
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

	DescribeTable("should return egress fromRules for the given external service",
		func(given testCase) {
			// given external service resource
			es := readES(given.esFile)
			// given policies
			resources, _ := readPolicies(given.policiesFile)

			// when
			policies, err := matchers.EgressMatchedPolicies(mtp_api.MeshTrafficPermissionType, es.Spec.Tags, resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(policies.FromRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		}, generateTableEntries(filepath.Join("testdata", "egressmatchedpolicies", "fromrules")))

	DescribeTable("should return egress fromRules for the given external service when policy has From and To",
		func(given testCase) {
			// given external service resource
			es := readES(given.esFile)
			// given policies
			resources, _ := readPolicies(given.policiesFile)

			// when
			policies, err := matchers.EgressMatchedPolicies(mt_api.MeshTimeoutType, es.Spec.Tags, resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(policies.FromRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		}, generateTableEntries(filepath.Join("testdata", "egressmatchedpolicies", "fromtorules")))

	DescribeTable("should return egress toRules for the given external service",
		func(given testCase) {
			// given external service resource
			es := readES(given.esFile)
			// given policies
			resources, _ := readPolicies(given.policiesFile)

			// when
			policies, err := matchers.EgressMatchedPolicies(v1alpha1.MeshLoadBalancingStrategyType, es.Spec.Tags, resources)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(policies.FromRules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(test_matchers.MatchGoldenYAML(given.goldenFile))
		}, generateTableEntries(filepath.Join("testdata", "egressmatchedpolicies", "torules")))

	DescribeTableSubtree("MeshExternalService ResourceRules",
		func(inputFile string) {
			DescribeTable("should build a rule-based view for to policies",
				func() {
					// given
					resources := file.ReadInputFile(inputFile)
					meshCtx := xds_builders.Context().WithMeshLocalResources(resources).Build()

					// when
					rules, err := matchers.EgressMatchedPolicies(mhc_api.MeshHealthCheckType, map[string]string{}, meshCtx.Mesh.Resources)
					Expect(err).ToNot(HaveOccurred())

					// then
					bytes, err := yaml.Marshal(rules)
					Expect(err).ToNot(HaveOccurred())
					Expect(bytes).To(test_matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.", ".golden.", 1)))
				},
				Entry("should generate to resource rules for egress and mesh externalservice"),
			)
		},
		test.EntriesForFolder(filepath.Join("egressmatchedpolicies", "meshexternalservice", "torules")),
	)

	DescribeTableSubtree("MeshExternalService ResourceRules from policy",
		func(inputFile string) {
			DescribeTable("should build a rule-based view for policies",
				func() {
					// given
					resources := file.ReadInputFile(inputFile)
					meshCtx := xds_builders.Context().WithMeshLocalResources(resources).Build()

					// when
					rules, err := matchers.EgressMatchedPolicies(mcb_api.MeshCircuitBreakerType, map[string]string{}, meshCtx.Mesh.Resources)
					Expect(err).ToNot(HaveOccurred())

					// then
					bytes, err := yaml.Marshal(rules)
					Expect(err).ToNot(HaveOccurred())
					Expect(bytes).To(test_matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.", ".golden.", 1)))
				},
				Entry("should generate to resource rules for egress and mesh externalservice"),
			)
		},
		test.EntriesForFolder(filepath.Join("egressmatchedpolicies", "meshexternalservice", "fromtorules")),
	)
})
