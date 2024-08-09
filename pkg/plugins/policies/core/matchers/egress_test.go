package matchers_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/plugins/policies/core/matchers"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	mt_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	mtp_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	test_matchers "github.com/kumahq/kuma/pkg/test/matchers"
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
})
