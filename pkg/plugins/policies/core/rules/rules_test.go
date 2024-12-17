package rules_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/file"
	"github.com/kumahq/kuma/pkg/xds/context"
)

var _ = Describe("Rules", func() {
	Describe("SubsetIter", func() {
		It("should return all possible subsets for the given set of tags", func() {
			// given
			tags := []core_rules.Tag{
				{Key: "k1", Value: "v1"},
				{Key: "k2", Value: "v2"},
				{Key: "k3", Value: "v3"},
			}

			// when
			iter := core_rules.NewSubsetIter(tags)

			// then
			expected := [][]core_rules.Tag{
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
			tags := []core_rules.Tag{}

			// when
			iter := core_rules.NewSubsetIter(tags)

			// then
			empty := iter.Next()
			Expect(empty).To(Equal(core_rules.Subset{}))
		})

		It("should handle tags with equal keys", func() {
			// given
			tags := []core_rules.Tag{
				{Key: "zone", Value: "us-east"},
				{Key: "env", Value: "dev"},
				{Key: "env", Value: "prod"},
			}

			// when
			iter := core_rules.NewSubsetIter(tags)

			// then
			expected := []core_rules.Subset{
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

	Describe("IsSubset", func() {
		type testCase struct {
			s1, s2   core_rules.Subset
			isSubset bool
		}

		DescribeTable("should respond if s2 is subset of s1",
			func(given testCase) {
				Expect(given.s1.IsSubset(given.s2)).To(Equal(given.isSubset))
			},
			Entry("entry 1", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "backend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "frontend"},
					{Key: "version", Value: "v2"},
				},
				isSubset: false,
			}),
			Entry("entry 2", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "backend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Value: "backend"},
					{Key: "version", Value: "v2"},
				},
				isSubset: true,
			}),
			Entry("entry 3", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Value: "v2"},
				},
				isSubset: true,
			}),
			Entry("entry 4", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				isSubset: true,
			}),
			Entry("entry 5", testCase{
				s1: []core_rules.Tag{},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				isSubset: true,
			}),
			Entry("entry 6", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "backend"},
					{Key: "version", Value: "v1"},
				},
				s2:       []core_rules.Tag{},
				isSubset: false,
			}),
			Entry("entry 7", testCase{
				s1: []core_rules.Tag{
					{Key: "key1", Not: true, Value: "val1"},
				},
				s2: []core_rules.Tag{
					{Key: "key1", Value: "val2"},
				},
				isSubset: true,
			}),
			Entry("entry 8", testCase{
				s1: []core_rules.Tag{
					{Key: "key1", Not: true, Value: "val1"},
				},
				s2: []core_rules.Tag{
					{Key: "key1", Value: "val2"},
					{Key: "key2", Value: "val3"},
				},
				isSubset: true,
			}),
		)
	})

	Describe("Intersect", func() {
		type testCase struct {
			s1, s2    core_rules.Subset
			intersect bool
		}

		DescribeTable("should respond if s1 and s2 have intersection",
			func(given testCase) {
				Expect(given.s1.Intersect(given.s2)).To(Equal(given.intersect))
			},
			Entry("entry 1", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "backend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "frontend"},
					{Key: "version", Value: "v2"},
				},
				intersect: true,
			}),
			Entry("entry 2", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "backend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Value: "frontend"},
					{Key: "version", Value: "v2"},
				},
				intersect: false,
			}),
			Entry("entry 3", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Value: "v2"},
				},
				intersect: true,
			}),
			Entry("entry 4", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("entry 5", testCase{
				s1: []core_rules.Tag{},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("entry 6", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2:        []core_rules.Tag{},
				intersect: true,
			}),
			Entry("entry 7", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Not: true, Value: "frontend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("entry 8", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Value: "frontend"},
				},
				intersect: true,
			}),
			Entry("entry 9", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []core_rules.Tag{
					{Key: "service", Value: "backend"},
				},
				intersect: false,
			}),
			Entry("entry 10", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []core_rules.Tag{
					{Key: "version", Value: "v1"},
				},
				intersect: true,
			}),
			Entry("entry 11", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "backend"},
					{Key: "version", Value: "v1"},
				},
				s2: []core_rules.Tag{
					{Key: "version", Value: "v2"},
					{Key: "zone", Value: "east"},
				},
				intersect: false,
			}),
			Entry("entry 12", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []core_rules.Tag{
					{Key: "version", Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("entry 13", testCase{
				s1: []core_rules.Tag{
					{Key: "service", Value: "frontend"},
					{Key: "version", Value: "v1"},
				},
				s2: []core_rules.Tag{
					{Key: "version", Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
		)
	})

	Describe("BuildRules", func() {
		buildRulesTestTemplate := func(inputFile string, fn func(policies []core_model.Resource) (interface{}, error)) {
			// given
			policies := file.ReadInputFile(inputFile)
			// when
			rules, err := fn(policies)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(rules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.yaml", ".golden.yaml", 1)))
		}

		DescribeTable("should build a rule-based view for the policy with a from list",
			func(inputFile string) {
				buildRulesTestTemplate(inputFile, func(policies []core_model.Resource) (interface{}, error) {
					// given
					listener := core_rules.InboundListener{
						Address: "127.0.0.1",
						Port:    80,
					}
					policiesByInbound := map[core_rules.InboundListener][]core_model.Resource{
						listener: policies,
					}
					// when
					return core_rules.BuildFromRules(policiesByInbound)
				})
			},
			test.EntriesForFolder("rules/from"),
		)

		DescribeTable("should build a rule-based view for the policy with a to list",
			func(inputFile string) {
				buildRulesTestTemplate(inputFile, func(policies []core_model.Resource) (interface{}, error) {
					actualPolicies := []core_model.Resource{}
					var httpRoutes []*v1alpha1.MeshHTTPRouteResource
					for _, policy := range policies {
						switch policy.Descriptor().Name {
						case v1alpha1.MeshHTTPRouteType:
							httpRoutes = append(httpRoutes, policy.(*v1alpha1.MeshHTTPRouteResource))
						default:
							actualPolicies = append(actualPolicies, policy)
						}
					}
					return core_rules.BuildToRules(actualPolicies, context.Resources{MeshLocalResources: map[core_model.ResourceType]core_model.ResourceList{
						v1alpha1.MeshHTTPRouteType: &v1alpha1.MeshHTTPRouteResourceList{Items: httpRoutes},
					}})
				})
			},
			test.EntriesForFolder("rules/to"),
		)

		DescribeTable("should build a rule-based view for list of single item policies",
			func(inputFile string) {
				buildRulesTestTemplate(inputFile, func(policies []core_model.Resource) (interface{}, error) {
					return core_rules.BuildSingleItemRules(policies)
				})
			},
			test.EntriesForFolder("rules/single"),
		)
	})

	Describe("Eval", func() {
		type testCase struct {
			rules    core_rules.Rules
			subset   core_rules.Subset
			confYAML []byte
		}

		DescribeTable("should compute conf for subset based on rules",
			func(given testCase) {
				element := core_rules.Element{}
				for _, tag := range given.subset {
					element[tag.Key] = tag.Value
				}

				conf := given.rules.NewCompute(element)
				if given.confYAML == nil {
					Expect(conf).To(BeNil())
				} else {
					actualYAML, err := yaml.Marshal(conf.Conf)
					Expect(err).To(Not(HaveOccurred()))
					Expect(actualYAML).To(MatchYAML(given.confYAML))
				}
			},
			Entry("single matched rule", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("single matched rule and subset", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val1"}, // rule has "key1: val1"
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("single matched not", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val2"},
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("single matched rule, rule and subset with negation", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val1"},
				},
				confYAML: nil,
			}),
			Entry("empty set is a superset for all subset", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{}, // empty set
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("no rules matched, rule with negation, subset without key", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key2", Value: "val2"},
				},
				confYAML: nil,
			}),
			Entry("no rules matched, subset has key which is not presented in superset", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key2", Value: "val2"}, // key2 is not in rules[0].Subset
				},
				confYAML: nil,
			}),
			Entry("no rules matched, subset has key with another value", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val2"}, // val2 is not equal to rules[0].Subset["key1"]
				},
				confYAML: nil,
			}),
			Entry("no rules matched, subset has same key and value", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val1"},
				},
				confYAML: nil,
			}),
			Entry("the first matched conf is taken", testCase{
				rules: core_rules.Rules{
					{
						Subset: core_rules.Subset{
							{Key: "key1", Value: "val1"}, // not matched
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
					{
						Subset: core_rules.Subset{
							{Key: "key2", Value: "val2"}, // the first matched
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Deny",
						},
					},
					{
						Subset: core_rules.Subset{}, // matched but not the first
						Conf: meshtrafficpermission_api.Conf{
							Action: "AllowWithShadowDeny",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key2", Value: "val2"},
					{Key: "key3", Value: "val3"},
				},
				confYAML: []byte(`action: Deny`),
			}),
			Entry("n dimensions rules and n-1 dimensions subsets", testCase{
				rules: core_rules.Rules{
					{
						Subset: []core_rules.Tag{
							{Key: "key1", Value: "val1"},
							{Key: "key2", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				subset: []core_rules.Tag{
					{Key: "key1", Value: "val1"},
				},
				confYAML: []byte(`action: Allow`),
			}),
		)
	})
})
