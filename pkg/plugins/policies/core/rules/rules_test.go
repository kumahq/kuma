package rules_test

import (
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/core/resources/registry"
	core_rules "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v2/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/v2/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	"github.com/kumahq/kuma/v2/pkg/test/resources/file"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
	"github.com/kumahq/kuma/v2/pkg/xds/context"
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
				{
					{Key: "zone", Value: "us-east"},
					{Key: "env", Value: "prod"},
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
						},
					},
					{
						Subset: subsetutils.Subset{
							{Key: "key2", Value: "val2"}, // the first matched
						},
						Conf: meshtrafficpermission_api.Conf{
							Action: pointer.To[meshtrafficpermission_api.Action]("Deny"),
						},
					},
					{
						Subset: subsetutils.Subset{}, // matched but not the first
						Conf: meshtrafficpermission_api.Conf{
							Action: pointer.To[meshtrafficpermission_api.Action]("AllowWithShadowDeny"),
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
							Action: pointer.To[meshtrafficpermission_api.Action]("Allow"),
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

	Describe("Clique vs Connected Components optimization", func() {
		// Helper to create a PolicyItemWithMeta for testing
		createPolicyItem := func(targetRef common_api.TargetRef, action meshtrafficpermission_api.Action) core_rules.PolicyItemWithMeta {
			return core_rules.PolicyItemWithMeta{
				PolicyItem: &meshtrafficpermission_api.From{
					TargetRef: targetRef,
					Default: meshtrafficpermission_api.Conf{
						Action: pointer.To(action),
					},
				},
				ResourceMeta: &test_model.ResourceMeta{Name: "test-policy", Mesh: "default"},
				TopLevel:     common_api.TargetRef{Kind: common_api.Mesh},
			}
		}

		// Helper to verify semantic equivalence: both approaches should return
		// the same configuration for any given element
		verifySemanticEquivalence := func(rulesCliques, rulesComponents core_rules.Rules, elements []subsetutils.Element) {
			for _, elem := range elements {
				confCliques := rulesCliques.Compute(elem)
				confComponents := rulesComponents.Compute(elem)

				if confCliques == nil && confComponents == nil {
					continue
				}
				Expect(confCliques).ToNot(BeNil(), "cliques rule should match element %v", elem)
				Expect(confComponents).ToNot(BeNil(), "components rule should match element %v", elem)
				Expect(confCliques.Conf).To(Equal(confComponents.Conf),
					"configuration should be the same for element %v", elem)
			}
		}

		It("should produce semantically equivalent results for disjoint subsets connected via Mesh", func() {
			// Mesh, {app: app-1}, {app: app-2}, {app: app-3}
			// All intersect with Mesh but not with each other
			// Cliques: [Mesh, app-1], [Mesh, app-2], [Mesh, app-3]
			// Connected components: [Mesh, app-1, app-2, app-3]
			// Note: cliques may produce more rules due to overlapping cliques being processed independently
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.Mesh,
				}, "AllowWithShadowDeny"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-1"},
				}, "Allow"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-2"},
				}, "Deny"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-3"},
				}, "Allow"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			// Both should produce valid rules
			Expect(rulesCliques).ToNot(BeEmpty())
			Expect(rulesComponents).ToNot(BeEmpty())

			// Verify semantic equivalence - this is the key property
			testElements := []subsetutils.Element{
				{"app": "app-1"},
				{"app": "app-2"},
				{"app": "app-3"},
				{"app": "app-4"}, // not in any specific rule, should match Mesh
				{"app": "app-1", "version": "v1"},
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})

		It("should handle overlapping subsets that form actual cliques", func() {
			// Subsets that all intersect with each other should form one clique
			// {zone: us-east}, {zone: us-east, env: prod}, {env: prod}
			// All three intersect with each other
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"zone": "us-east"},
				}, "Allow"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"zone": "us-east", "env": "prod"},
				}, "Deny"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"env": "prod"},
				}, "AllowWithShadowDeny"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			// For fully connected subsets, both approaches should produce similar results
			testElements := []subsetutils.Element{
				{"zone": "us-east", "env": "prod"},
				{"zone": "us-east", "env": "dev"},
				{"zone": "us-west", "env": "prod"},
				{"zone": "us-west", "env": "dev"},
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})

		It("should handle chain of subsets (A-B-C where A and C don't intersect)", func() {
			// A: {app: app-1, version: v1}
			// B: {app: app-1}  - intersects with both A and C
			// C: {app: app-1, version: v2}
			// A and C don't intersect because version=v1 and version=v2 are mutually exclusive
			// But they're connected via B in the graph
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-1", "version": "v1"},
				}, "Allow"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-1"},
				}, "Deny"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-1", "version": "v2"},
				}, "AllowWithShadowDeny"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			testElements := []subsetutils.Element{
				{"app": "app-1", "version": "v1"},
				{"app": "app-1", "version": "v2"},
				{"app": "app-1", "version": "v3"}, // matches only B
				{"app": "app-1"},
				{"app": "app-2"},
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})

		It("should handle empty input", func() {
			items := []core_rules.PolicyItemWithMeta{}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())
			Expect(rulesCliques).To(BeEmpty())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())
			Expect(rulesComponents).To(BeEmpty())
		})

		It("should handle single subset", func() {
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{Kind: common_api.Mesh}, "Allow"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			verifySemanticEquivalence(rulesCliques, rulesComponents, []subsetutils.Element{
				{"kuma.io/service": "any"},
			})
		})

		It("should handle identical subsets", func() {
			// Multiple policies with the same targetRef
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-1"},
				}, "Allow"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-1"},
				}, "Deny"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			testElements := []subsetutils.Element{
				{"app": "app-1"},
				{"app": "app-2"},
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})

		It("should demonstrate efficiency gain with many disjoint subsets", func() {
			// This simulates the mtp-long test case
			// Many subsets with unique tag combinations that don't intersect with each other
			// but all intersect with Mesh (empty set)
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{Kind: common_api.Mesh}, "AllowWithShadowDeny"),
			}
			for i := 1; i <= 5; i++ {
				items = append(items, createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{
						"app":       fmt.Sprintf("app-%d", i),
						"namespace": fmt.Sprintf("ns-%d", i),
					},
				}, "Allow"))
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			// With disjoint subsets connected only through Mesh, cliques produce significantly fewer rules.
			// Connected components puts all nodes in one component (connected via Mesh), creating
			// combinatorial explosion with 5 app values × 5 namespace values.
			// Cliques separates them into smaller groups: [Mesh, app-1/ns-1], [Mesh, app-2/ns-2], etc.
			Expect(len(rulesCliques)).To(BeNumerically("<", len(rulesComponents)),
				"cliques should produce fewer rules for disjoint subsets: cliques=%d, components=%d",
				len(rulesCliques), len(rulesComponents))

			// Still semantically equivalent
			testElements := []subsetutils.Element{
				{"app": "app-1", "namespace": "ns-1"},
				{"app": "app-2", "namespace": "ns-2"},
				{"app": "app-1", "namespace": "ns-2"}, // mix of tags - should match Mesh
				{"app": "other"},                      // should match Mesh
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})

		It("should handle completely disjoint subsets (no edges in graph)", func() {
			// Subsets that don't intersect at all
			// {app: app-1} and {app: app-2} don't intersect because app can only have one value
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-1"},
				}, "Allow"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"app": "app-2"},
				}, "Deny"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			// Both should handle disjoint subsets correctly
			testElements := []subsetutils.Element{
				{"app": "app-1"},
				{"app": "app-2"},
				{"app": "app-3"},
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})

		It("should handle MeshSubset with multiple tags", func() {
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{Kind: common_api.Mesh}, "AllowWithShadowDeny"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"zone": "us-east", "env": "prod", "team": "platform"},
				}, "Allow"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"zone": "us-west", "env": "dev", "team": "product"},
				}, "Deny"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			testElements := []subsetutils.Element{
				{"zone": "us-east", "env": "prod", "team": "platform"},
				{"zone": "us-west", "env": "dev", "team": "product"},
				{"zone": "us-east", "env": "dev"},
				{"zone": "other"},
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})

		It("should handle partial overlap between subsets", func() {
			// A: {zone: us-east}
			// B: {zone: us-east, env: prod}  - subset of A, intersects with C
			// C: {env: prod}  - intersects with B but not subset of A
			// D: {env: dev}  - disjoint from B and C, intersects with A via Mesh-like behavior
			items := []core_rules.PolicyItemWithMeta{
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"zone": "us-east"},
				}, "Allow"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"zone": "us-east", "env": "prod"},
				}, "Deny"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"env": "prod"},
				}, "AllowWithShadowDeny"),
				createPolicyItem(common_api.TargetRef{
					Kind: common_api.MeshSubset,
					Tags: &map[string]string{"env": "dev"},
				}, "Allow"),
			}

			rulesCliques, err := core_rules.BuildRules(items, true, true)
			Expect(err).ToNot(HaveOccurred())

			rulesComponents, err := core_rules.BuildRules(items, true, false)
			Expect(err).ToNot(HaveOccurred())

			testElements := []subsetutils.Element{
				{"zone": "us-east", "env": "prod"},
				{"zone": "us-east", "env": "dev"},
				{"zone": "us-west", "env": "prod"},
				{"zone": "us-west", "env": "dev"},
				{"zone": "us-east"},
				{"env": "prod"},
				{"env": "dev"},
			}
			verifySemanticEquivalence(rulesCliques, rulesComponents, testElements)
		})
	})
})
