package rules_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
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
			tags := []subsetutils.Tag{
				{Key: "k1", Value: "v1"},
				{Key: "k2", Value: "v2"},
				{Key: "k3", Value: "v3"},
			}

			// when
			iter := subsetutils.NewSubsetIter(tags)

			// then
			expected := [][]subsetutils.Tag{
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
			tags := []subsetutils.Tag{}

			// when
			iter := subsetutils.NewSubsetIter(tags)

			// then
			empty := iter.Next()
			Expect(empty).To(Equal(subsetutils.Subset{}))
		})

		It("should handle tags with equal keys", func() {
			// given
			tags := []subsetutils.Tag{
				{Key: "zone", Value: "us-east"},
				{Key: "env", Value: "dev"},
				{Key: "env", Value: "prod"},
			}

			// when
			iter := subsetutils.NewSubsetIter(tags)

			// then
			expected := []subsetutils.Subset{
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
			s1, s2   subsetutils.Subset
			isSubset bool
		}

		DescribeTable("should respond if s2 is subset of s1",
			func(given testCase) {
				Expect(given.s1.IsSubset(given.s2)).To(Equal(given.isSubset))
			},
			Entry("entry 1", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
					{Key: "version", Value: "v2"},
				},
				isSubset: false,
			}),
			Entry("entry 2", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
					{Key: "version", Value: "v2"},
				},
				isSubset: true,
			}),
			Entry("entry 3", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Value: "v2"},
				},
				isSubset: true,
			}),
			Entry("entry 4", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				isSubset: true,
			}),
			Entry("entry 5", testCase{
				s1: []subsetutils.Tag{},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				isSubset: true,
			}),
			Entry("entry 6", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
					{Key: "version", Value: "v1"},
				},
				s2:       []subsetutils.Tag{},
				isSubset: false,
			}),
			Entry("entry 7", testCase{
				s1: []subsetutils.Tag{
					{Key: "key1", Not: true, Value: "val1"},
				},
				s2: []subsetutils.Tag{
					{Key: "key1", Value: "val2"},
				},
				isSubset: true,
			}),
			Entry("entry 8", testCase{
				s1: []subsetutils.Tag{
					{Key: "key1", Not: true, Value: "val1"},
				},
				s2: []subsetutils.Tag{
					{Key: "key1", Value: "val2"},
					{Key: "key2", Value: "val3"},
				},
				isSubset: true,
			}),
		)
	})

	Describe("Intersect", func() {
		type testCase struct {
			s1, s2    subsetutils.Subset
			intersect bool
		}

		DescribeTable("should respond if s1 and s2 have intersection",
			func(given testCase) {
				Expect(given.s1.Intersect(given.s2)).To(Equal(given.intersect))
			},
			Entry("positive, same key and value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				intersect: true,
			}),
			Entry("positive, same key, different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
				},
				intersect: false,
			}),
			Entry("positive, multiple key-values with overlap key-value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
					{Key: "version", Value: "v1"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				intersect: true,
			}),
			Entry("positive, multiple key-values with overlap key but different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
					{Key: "version", Value: "v1"},
				},
				intersect: false,
			}),
			Entry("positive, different key, different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "version", Value: "v1"},
				},
				intersect: true,
			}),
			Entry("positive, superset", testCase{
				s1: []subsetutils.Tag{},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
					{Key: "version", Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("a part of negation, same key and value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				intersect: true,
			}),
			Entry("a part of negation, same key, different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
				},
				intersect: true,
			}),
			Entry("a part of negation, multiple key-values with overlap key-value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
					{Key: "version", Value: "v1"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "frontend"},
				},
				intersect: true,
			}),
			Entry("a part of negation, multiple key-values with overlap key but different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
					{Key: "version", Value: "v1"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Value: "backend"},
				},
				intersect: true,
			}),
			Entry("a part of negation, different key, different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "version", Value: "v1"},
				},
				intersect: true,
			}),
			Entry("negation, same key and value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
				},
				intersect: true,
			}),
			Entry("negation, same key, different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
				},
				intersect: true,
			}),
			Entry("negation, multiple key-values with overlap key-value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("negation, multiple key-values with overlap key but different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("negation, multiple key-values with overlap key but different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("negation, different key, different value", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "frontend"},
				},
				s2: []subsetutils.Tag{
					{Key: "version", Not: true, Value: "v1"},
					{Key: "zone", Value: "east"},
				},
				intersect: true,
			}),
			Entry("negation, superset", testCase{
				s1: []subsetutils.Tag{
					{Key: "service", Not: true, Value: "backend"},
					{Key: "version", Not: true, Value: "v1"},
				},
				s2:        []subsetutils.Tag{},
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

		samePolicyTypesToList := func(policies []core_model.Resource) core_model.ResourceList {
			Expect(policies).ToNot(BeEmpty())
			policyType := policies[0].Descriptor().Name
			list, err := registry.Global().NewList(policyType)
			Expect(err).ToNot(HaveOccurred())
			for _, policy := range policies {
				Expect(list.AddItem(policy)).To(Succeed())
			}
			return list
		}

		DescribeTable("should build a rule-based view for the policy with a from list",
			func(inputFile string) {
				buildRulesTestTemplate(inputFile, func(policies []core_model.Resource) (interface{}, error) {
					// given
					listener := core_rules.InboundListener{
						Address: "127.0.0.1",
						Port:    80,
					}
					policiesByInbound := map[core_rules.InboundListener]core_model.ResourceList{
						listener: samePolicyTypesToList(policies),
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
					var actualPolicies core_model.ResourceList
					var httpRoutes []*v1alpha1.MeshHTTPRouteResource
					for _, policy := range policies {
						switch policy.Descriptor().Name {
						case v1alpha1.MeshHTTPRouteType:
							httpRoutes = append(httpRoutes, policy.(*v1alpha1.MeshHTTPRouteResource))
						default:
							if actualPolicies == nil {
								var err error
								actualPolicies, err = registry.Global().NewList(policy.Descriptor().Name)
								Expect(err).ToNot(HaveOccurred())
							}
							Expect(actualPolicies.AddItem(policy)).To(Succeed())
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

	Describe("ContainsElement", func() {
		type testCase struct {
			ss       subsetutils.Subset
			other    subsetutils.Element
			contains bool
		}

		DescribeTable("should respond if subset ss contains element other",
			func(given testCase) {
				Expect(given.ss.ContainsElement(given.other)).To(Equal(given.contains))
			},
			Entry("single matched rule by single rule and elements", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
				},
				other: subsetutils.Element{
					"key1": "val1",
					"key2": "val2",
				},
				contains: true,
			}),
			Entry("single matched rule by single rule and element", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
				},
				other: subsetutils.Element{
					"key1": "val1",
				},
				contains: true,
			}),
			Entry("single matched rule, rule with negation, element has key with another value", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1", Not: true},
				},
				other: subsetutils.Element{
					"key1": "val2",
				},
				contains: true,
			}),
			Entry("empty set is a superset for all element", testCase{
				ss: []subsetutils.Tag{},
				other: subsetutils.Element{
					"key1": "val2",
				},
				contains: true,
			}),
			Entry("empty element", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
				},
				other:    subsetutils.Element{},
				contains: false,
			}),
			Entry("no rules matched, rule with negation, element has same key value", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1", Not: true},
				},
				other: subsetutils.Element{
					"key1": "val1",
				},
				contains: false,
			}),
			Entry("no rules matched, rule with negation, element has another key", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1", Not: true},
				},
				other: subsetutils.Element{
					"key2": "val2",
				},
				contains: true,
			}),
			Entry("no rules matched, element has key which is not presented in superset", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
				},
				other: subsetutils.Element{
					"key2": "val2",
				},
				contains: false,
			}),

			Entry("no rules matched, rules with positive, element has key with another value", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
				},
				other: subsetutils.Element{
					"key1": "val2",
				},
				contains: false,
			}),
			Entry("no rules matched, rules with positive, element has only one overlapped key value", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				other: subsetutils.Element{
					"key1": "val1",
				},
				contains: false,
			}),
			Entry("single matched rule by rules and element, rules with a part of negation", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2", Not: true},
				},
				other: subsetutils.Element{
					"key1": "val1",
				},
				contains: true,
			}),
			Entry("single matched rule by rules and element, rules with a part of negation, element has key with another value", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1", Not: true},
					{Key: "key2", Value: "val2"},
				},
				other: subsetutils.Element{
					"key1": "val2",
				},
				contains: false,
			}),
			Entry("no rules matched, rules with negation, element has same key value", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1", Not: true},
					{Key: "key2", Value: "val2", Not: true},
				},
				other: subsetutils.Element{
					"key1": "val1",
				},
				contains: false,
			}),
			Entry("no rules matched, rules with a part of negation, element has another key", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2", Not: true},
				},
				other: subsetutils.Element{
					"key3": "val3",
				},
				contains: false,
			}),
			Entry("no rules matched, rules with positive, element has another key", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				other: subsetutils.Element{
					"key3": "val3",
				},
				contains: false,
			}),
			Entry("rules matched, rules with negation, element has another key", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1", Not: true},
					{Key: "key2", Value: "val2", Not: true},
				},
				other: subsetutils.Element{
					"key3": "val3",
				},
				contains: true,
			}),
			Entry("no rules matched, n dimensions rules and n-1 dimensions elements, rules with positive, elements have another keys", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
					{Key: "key3", Value: "val3"},
				},
				other: subsetutils.Element{
					"key4": "val4",
					"key5": "val5",
				},
				contains: false,
			}),
			Entry("no rules matched, n dimensions rules and n-1 dimensions elements, rules with positive, elements have overlapped key by rules", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
					{Key: "key3", Value: "val3"},
				},
				other: subsetutils.Element{
					"key3": "val3",
					"key4": "val4",
				},
				contains: false,
			}),
			Entry("rules matched, n dimensions rules and n-1 dimensions elements, rules with a part of negation, elements have overlapped key by rules", testCase{
				ss: []subsetutils.Tag{
					{Key: "key1", Value: "val1", Not: true},
					{Key: "key2", Value: "val2", Not: true},
					{Key: "key3", Value: "val3"},
				},
				other: subsetutils.Element{
					"key3": "val3",
					"key4": "val4",
				},
				contains: true,
			}),
		)
	})

	Describe("Eval", func() {
		type testCase struct {
			rules    core_rules.Rules
			element  subsetutils.Element
			confYAML []byte
		}

		DescribeTable("should compute conf for subset based on rules",
			func(given testCase) {
				conf := given.rules.Compute(given.element)
				if given.confYAML == nil {
					Expect(conf).To(BeNil())
				} else {
					actualYAML, err := yaml.Marshal(conf.Conf)
					Expect(err).To(Not(HaveOccurred()))
					Expect(actualYAML).To(MatchYAML(given.confYAML))
				}
			},
			Entry("single matched rule by single rule and elements", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key1": "val1",
					"key2": "val2",
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("single matched rule by single rule and element", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key1": "val1",
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("single matched rule, rule with negation, element has key with another value", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key1": "val2",
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("empty set is a superset for all element", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{}, // empty set
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key1": "val1",
					"key2": "val2",
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("empty element", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element:  subsetutils.Element{},
				confYAML: nil,
			}),
			Entry("no rules matched, rule with negation, element has same key value", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key1": "val1",
				},
				confYAML: nil,
			}),
			Entry("no rules matched, rule with negation, element has another key", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key2": "val2",
				},
				confYAML: []byte(`action: Allow`),
			}),
			Entry("no rules matched, element has key which is not presented in superset", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key2": "val2", // key2 is not in rules[0].Subset
				},
				confYAML: nil,
			}),
			Entry("no rules matched, element has key with another value", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1"},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key1": "val2", // val2 is not equal to rules[0].Subset["key1"]
				},
				confYAML: nil,
			}),
			Entry("the first matched conf is taken", testCase{
				rules: core_rules.Rules{
					{
						Subset: subsetutils.Subset{
							{Key: "key1", Value: "val1"}, // not matched
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
					{
						Subset: subsetutils.Subset{
							{Key: "key2", Value: "val2"}, // the first matched
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Deny",
						},
					},
					{
						Subset: subsetutils.Subset{}, // matched but not the first
						Conf: meshtrafficpermission_api.Conf{
							Action: "AllowWithShadowDeny",
						},
					},
				},
				element: subsetutils.Element{
					"key2": "val2",
					"key3": "val3",
				},
				confYAML: []byte(`action: Deny`),
			}),
			Entry("n dimensions subset and n-1 dimensions elements", testCase{
				rules: core_rules.Rules{
					{
						Subset: []subsetutils.Tag{
							{Key: "key1", Value: "val1"},
							{Key: "key2", Value: "val1", Not: true},
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: "Allow",
						},
					},
				},
				element: subsetutils.Element{
					"key1": "val1",
				},
				confYAML: []byte(`action: Allow`),
			}),
		)
	})
})
