package xds_test

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/xds"
	_ "github.com/kumahq/kuma/pkg/plugins/policies"
	meshtrace_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrace/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test/matchers"
	util_yaml "github.com/kumahq/kuma/pkg/util/yaml"
)

var _ = Describe("Rules", func() {

	Describe("SubsetIter", func() {

		It("should return all possible subsets for the given set of tags", func() {
			// given
			tags := []xds.Tag{
				{Key: "k1", Value: "v1"},
				{Key: "k2", Value: "v2"},
				{Key: "k3", Value: "v3"},
			}

			// when
			iter := xds.NewSubsetIter(tags)

			// then
			expected := [][]xds.Tag{
				{
					{Key: "k1", Not: true, Value: "v1"},
					{Key: "k2", Value: "v2"},
					{Key: "k3", Value: "v3"},
				},
				{
					{Key: "k1", Value: "v1"},
					{Key: "k2", Not: true, Value: "v2"},
					{Key: "k3", Value: "v3"},
				},
				{
					{Key: "k1", Not: true, Value: "v1"},
					{Key: "k2", Not: true, Value: "v2"},
					{Key: "k3", Value: "v3"},
				},
				{
					{Key: "k1", Value: "v1"},
					{Key: "k2", Value: "v2"},
					{Key: "k3", Not: true, Value: "v3"},
				},
				{
					{Key: "k1", Not: true, Value: "v1"},
					{Key: "k2", Value: "v2"},
					{Key: "k3", Not: true, Value: "v3"},
				},
				{
					{Key: "k1", Value: "v1"},
					{Key: "k2", Not: true, Value: "v2"},
					{Key: "k3", Not: true, Value: "v3"},
				},
				{
					{Key: "k1", Not: true, Value: "v1"},
					{Key: "k2", Not: true, Value: "v2"},
					{Key: "k3", Not: true, Value: "v3"},
				},
				{
					{Key: "k1", Value: "v1"},
					{Key: "k2", Value: "v2"},
					{Key: "k3", Value: "v3"},
				},
			}
			for _, expectedTags := range expected {
				actual := iter.Next()
				Expect(actual).To(ConsistOf(expectedTags))
			}
			Expect(iter.Next()).To(BeNil())
		})

		It("should handle empty tags", func() {
			// given
			tags := []xds.Tag{}

			// when
			iter := xds.NewSubsetIter(tags)

			// then
			empty := iter.Next()
			Expect(empty).To(Equal(xds.Subset{}))
		})

		It("should handle tags with equal keys", func() {
			// given
			tags := []xds.Tag{
				{Key: "zone", Value: "us-east"},
				{Key: "env", Value: "dev"},
				{Key: "env", Value: "prod"},
			}

			// when
			iter := xds.NewSubsetIter(tags)

			// then
			expected := []xds.Subset{
				{
					{Key: "zone", Value: "us-east"},
					{Key: "env", Value: "prod"},
				},
				{
					{Key: "zone", Value: "us-east", Not: true},
					{Key: "env", Value: "prod"},
				},
				{
					{Key: "zone", Value: "us-east"},
					{Key: "env", Value: "dev"},
				},
				{
					{Key: "zone", Value: "us-east", Not: true},
					{Key: "env", Value: "dev"},
				},
				{
					{Key: "zone", Value: "us-east"},
					{Key: "env", Value: "dev", Not: true},
					{Key: "env", Value: "prod", Not: true},
				},
				{
					{Key: "zone", Value: "us-east", Not: true},
					{Key: "env", Value: "dev", Not: true},
					{Key: "env", Value: "prod", Not: true},
				},
			}
			for _, expectedTags := range expected {
				actual := iter.Next()
				Expect(actual).To(Equal(expectedTags))
			}
			Expect(iter.Next()).To(BeNil())
		})
	})

	Describe("BuildRules", func() {

		type testCase struct {
			policyFile string
			goldenFile string
		}

		DescribeTable("should build a rule-based view for the policy with a from list",
			func(given testCase) {
				// given
				policyBytes, err := os.ReadFile(path.Join("testdata", "rules", given.policyFile))
				Expect(err).ToNot(HaveOccurred())

				policy, err := rest.YAML.UnmarshalCore(policyBytes)
				Expect(err).ToNot(HaveOccurred())
				mtp, ok := policy.GetSpec().(xds.PolicyWithFromList)
				Expect(ok).To(BeTrue())

				// when
				rules, err := xds.BuildRules(xds.BuildPolicyItemsWithMeta(mtp.GetFromList(), policy.GetMeta()))
				Expect(err).ToNot(HaveOccurred())

				// then
				bytes, err := yaml.Marshal(rules)
				Expect(err).ToNot(HaveOccurred())

				Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "rules", given.goldenFile)))
			},
			Entry("01. MeshTrafficPermission with 2 'env' tags that have different values", testCase{
				policyFile: "01.policy.yaml",
				goldenFile: "01.golden.yaml",
			}),
			Entry("02. MeshTrafficPermission with 3 different tags", testCase{
				policyFile: "02.policy.yaml",
				goldenFile: "02.golden.yaml",
			}),
			Entry("03. MeshTrafficPermission with MeshService targets", testCase{
				policyFile: "03.policy.yaml",
				goldenFile: "03.golden.yaml",
			}),
			Entry("04. MeshAccessLog with overriding empty backend list", testCase{
				policyFile: "04.policy.yaml",
				goldenFile: "04.golden.yaml",
			}),
			Entry("05. MeshAccessLog with overriding list of different backend type", testCase{
				policyFile: "05.policy.yaml",
				goldenFile: "05.golden.yaml",
			}),
		)

		DescribeTable("should build a rule-based view for list of single item policies",
			func(given testCase) {
				// given
				policyBytes, err := os.ReadFile(path.Join("testdata", "rules", given.policyFile))
				Expect(err).ToNot(HaveOccurred())

				yamls := util_yaml.SplitYAML(string(policyBytes))
				policies := []xds.PolicyItemWithMeta{}
				for _, yaml := range yamls {
					policy, err := rest.YAML.UnmarshalCore([]byte(yaml))
					Expect(err).ToNot(HaveOccurred())
					mt, ok := policy.(*meshtrace_api.MeshTraceResource)
					Expect(ok).To(BeTrue())
					policies = append(policies, xds.PolicyItemWithMeta{mt.Spec.GetPolicyItem(), policy.GetMeta()})
				}

				// when
				rules, err := xds.BuildRules(policies)
				Expect(err).ToNot(HaveOccurred())

				// then
				bytes, err := yaml.Marshal(rules)
				Expect(err).ToNot(HaveOccurred())

				Expect(bytes).To(matchers.MatchGoldenYAML(path.Join("testdata", "rules", given.goldenFile)))
			},
			Entry("06. MeshTrace", testCase{
				policyFile: "06.policy.yaml",
				goldenFile: "06.golden.yaml",
			}),
			Entry("07. MeshTrace list", testCase{
				policyFile: "07.policy.yaml",
				goldenFile: "07.golden.yaml",
			}),
		)
	})

	Describe("Eval", func() {

		type testCase struct {
			rules    xds.Rules
			subset   xds.Subset
			confYAML []byte
		}

		DescribeTable("should compute conf for subset based on rules",
			func(given testCase) {
				conf := given.rules.Compute(given.subset)
				if given.confYAML == nil {
					Expect(conf).To(BeNil())
				} else {
					actualYAML, err := yaml.Marshal(conf.Conf)
					Expect(err).To(Not(HaveOccurred()))
					Expect(actualYAML).To(MatchYAML(given.confYAML))
				}
			},
			Entry("single matched rule", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				confYAML: []byte(`action: ALLOW`),
			}),
			Entry("single matched rule, rule and subset with negation", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key1", Value: "val1", Not: true},
				},
				confYAML: []byte(`action: ALLOW`),
			}),
			Entry("empty set is a superset for all subset", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{}, // empty set
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				confYAML: []byte(`action: ALLOW`),
			}),
			Entry("no rules matched, rule with negation, subset without key", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key2", Value: "val2"},
				},
				confYAML: nil,
			}),
			Entry("no rules matched, subset has key which is not presented in superset", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key2", Value: "val2"}, // key2 is not in rules[0].Subset
				},
				confYAML: nil,
			}),
			Entry("no rules matched, subset has key with another value", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key1", Value: "val2"}, // val2 is not equal to rules[0].Subset["key1"]
				},
				confYAML: nil,
			}),
			Entry("no rules matched, rule with negation", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key1", Value: "val1"}, // rule has "key1: !val1"
				},
				confYAML: nil,
			}),
			Entry("no rules matched, subset with negation", testCase{
				rules: xds.Rules{
					{
						Subset: []xds.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key1", Value: "val1", Not: true}, // rule has "key1: val1"
				},
				confYAML: nil,
			}),
			Entry("the first matched conf is taken", testCase{
				rules: xds.Rules{
					{
						Subset: xds.Subset{
							{Key: "key1", Value: "val1"}, // not matched
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW",
						},
					},
					{
						Subset: xds.Subset{
							{Key: "key2", Value: "val2"}, // the first matched
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "DENY",
						},
					},
					{
						Subset: xds.Subset{}, // matched but not the first
						Conf: meshtrafficpermission_api.Conf{
							Action: "ALLOW_WITH_SHADOW_DENY",
						},
					},
				},
				subset: []xds.Tag{
					{Key: "key2", Value: "val2"},
					{Key: "key3", Value: "val3"},
				},
				confYAML: []byte(`action: DENY`),
			}),
		)
	})
})
