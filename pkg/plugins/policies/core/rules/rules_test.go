package rules_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtrafficpermission_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/file"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
	"github.com/kumahq/kuma/v3/pkg/xds/context"
)

// toTestPolicy / toTestPolicyItem are minimal PolicyWithToList / PolicyItem
// stand-ins used to exercise buildToListWithRoutes without depending on a
// concrete policy type (whose validation forbids a top-level MeshHTTPRoute ref).
type toTestPolicy struct {
	targetRef common_api.TargetRef
	toList    []core_model.PolicyItem
}

func (p *toTestPolicy) GetTargetRef() common_api.TargetRef { return p.targetRef }
func (p *toTestPolicy) GetToList() []core_model.PolicyItem { return p.toList }

type toTestPolicyItem struct {
	targetRef common_api.TargetRef
	conf      any
}

func (t *toTestPolicyItem) GetTargetRef() common_api.TargetRef { return t.targetRef }
func (t *toTestPolicyItem) GetDefault() any                    { return t.conf }

var _ = Describe("Rules", func() {
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

	Describe("BuildRules", func() {
		buildRulesTestTemplate := func(inputFile string, fn func(policies []core_model.Resource) (any, error)) {
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

		It("should build inbound rules without legacy subset rules", func() {
			input, err := os.ReadFile("../matchers/testdata/matchedpolicies/fromrules/01.policies.yaml")
			Expect(err).ToNot(HaveOccurred())

			var policies []core_model.Resource
			for _, policyBytes := range util_yaml.SplitYAML(string(input)) {
				policy, err := rest.YAML.UnmarshalCore([]byte(policyBytes))
				Expect(err).ToNot(HaveOccurred())
				policies = append(policies, policy)
			}

			listener := core_rules.InboundListener{
				Address: "1.1.1.1",
				Port:    8080,
			}

			actual, err := core_rules.BuildFromRules(map[core_rules.InboundListener]core_model.ResourceList{
				listener: samePolicyTypesToList(policies),
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.InboundRules).To(HaveLen(1))
			Expect(actual.InboundRules).To(HaveKey(listener))
			Expect(actual.InboundRules[listener]).To(HaveLen(3))

			bytes, err := yaml.Marshal(actual.InboundRules[listener])
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(ContainSubstring("spiffe://mesh-1/backend"))
			Expect(string(bytes)).To(ContainSubstring("spiffe://mesh-1/orders"))
			Expect(string(bytes)).To(ContainSubstring("spiffe://mesh-1"))
		})

		DescribeTable("should build a rule-based view for the policy with a to list",
			func(inputFile string) {
				buildRulesTestTemplate(inputFile, func(policies []core_model.Resource) (any, error) {
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
				buildRulesTestTemplate(inputFile, func(policies []core_model.Resource) (any, error) {
					return core_rules.BuildProxyConf(policies)
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
})

var _ = Describe("buildToListWithRoutes", func() {
	route := func(name, namespace, backend string) core_model.Resource {
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
				To: &[]v1alpha1.To{{
					TargetRef: common_api.OutboundTargetRef{
						Kind:   common_api.OutboundTargetRefKindMeshService,
						Labels: pointer.To(map[string]string{mesh_proto.DisplayName: backend}),
					},
					Rules: []v1alpha1.Rule{{}},
				}},
			},
		}
	}

	policyTargetingRoute := &toTestPolicy{
		targetRef: common_api.TargetRef{
			Kind:   common_api.MeshHTTPRoute,
			Labels: pointer.To(map[string]string{mesh_proto.DisplayName: "route-1"}),
		},
		toList: []core_model.PolicyItem{
			&toTestPolicyItem{targetRef: common_api.TargetRef{Kind: common_api.Mesh}, conf: map[string]any{}},
		},
	}

	routes := []core_model.Resource{
		route("route-1.ns-a", "ns-a", "backend-a"),
		route("route-1.ns-b", "ns-b", "backend-b"),
	}

	expandedServices := func(items []core_model.PolicyItem) []string {
		var services []string
		for _, item := range items {
			services = append(services, pointer.Deref(item.GetTargetRef().Labels)[mesh_proto.DisplayName])
		}
		return services
	}

	It("expands only same-namespace routes for a namespaced (consumer) policy", func() {
		meta := &test_model.ResourceMeta{
			Mesh: "mesh-1",
			Name: "timeout-1",
			Labels: map[string]string{
				mesh_proto.PolicyRoleLabel:  string(mesh_proto.ConsumerPolicyRole),
				mesh_proto.KubeNamespaceTag: "ns-a",
			},
		}

		items, err := core_rules.BuildToListWithRoutesForTesting(meta, policyTargetingRoute, routes)
		Expect(err).ToNot(HaveOccurred())
		Expect(expandedServices(items)).To(ConsistOf("backend-a"))
	})

	It("expands all matching routes for a namespace-agnostic (system) policy", func() {
		meta := &test_model.ResourceMeta{Mesh: "mesh-1", Name: "timeout-1"}

		items, err := core_rules.BuildToListWithRoutesForTesting(meta, policyTargetingRoute, routes)
		Expect(err).ToNot(HaveOccurred())
		Expect(expandedServices(items)).To(ConsistOf("backend-a", "backend-b"))
	})

	It("fails when a route target selects a service by labels without kuma.io/display-name", func() {
		routeWithoutDisplayName := &v1alpha1.MeshHTTPRouteResource{
			Meta: &test_model.ResourceMeta{
				Mesh: "mesh-1",
				Name: "route-1.ns-a",
				Labels: map[string]string{
					mesh_proto.DisplayName:      "route-1",
					mesh_proto.KubeNamespaceTag: "ns-a",
				},
			},
			Spec: &v1alpha1.MeshHTTPRoute{
				To: &[]v1alpha1.To{{
					TargetRef: common_api.OutboundTargetRef{
						Kind:   common_api.OutboundTargetRefKindMeshService,
						Labels: pointer.To(map[string]string{"env": "dev"}),
					},
					Rules: []v1alpha1.Rule{{}},
				}},
			},
		}
		meta := &test_model.ResourceMeta{Mesh: "mesh-1", Name: "timeout-1"}

		_, err := core_rules.BuildToListWithRoutesForTesting(meta, policyTargetingRoute, []core_model.Resource{routeWithoutDisplayName})
		Expect(err).To(MatchError(ContainSubstring("kuma.io/display-name label is required")))
	})
})
