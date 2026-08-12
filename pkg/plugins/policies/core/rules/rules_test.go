package rules_test

import (
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/v3/pkg/core/resources/registry"
	core_rules "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v3/pkg/test"
	"github.com/kumahq/kuma/v3/pkg/test/matchers"
	"github.com/kumahq/kuma/v3/pkg/test/resources/file"
	util_yaml "github.com/kumahq/kuma/v3/pkg/util/yaml"
)

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
			Entry("positive subset", testCase{
				s1: []subsetutils.Tag{{Key: "service", Value: "backend"}},
				s2: []subsetutils.Tag{{Key: "service", Value: "backend"}, {Key: "version", Value: "v2"}},
				isSubset: true,
			}),
			Entry("negative subset", testCase{
				s1: []subsetutils.Tag{{Key: "service", Value: "backend"}},
				s2: []subsetutils.Tag{{Key: "service", Not: true, Value: "frontend"}, {Key: "version", Value: "v2"}},
				isSubset: false,
			}),
			Entry("negative selector can contain different value", testCase{
				s1: []subsetutils.Tag{{Key: "key1", Not: true, Value: "val1"}},
				s2: []subsetutils.Tag{{Key: "key1", Value: "val2"}},
				isSubset: true,
			}),
			Entry("empty superset", testCase{
				s1:       []subsetutils.Tag{},
				s2:       []subsetutils.Tag{{Key: "service", Value: "backend"}},
				isSubset: true,
			}),
		)
	})

	Describe("BuildRules", func() {
		samePolicyTypesToList := func(policies []model.Resource) model.ResourceList {
			Expect(policies).ToNot(BeEmpty())
			policyType := policies[0].Descriptor().Name
			list, err := registry.Global().NewList(policyType)
			Expect(err).ToNot(HaveOccurred())
			for _, policy := range policies {
				Expect(list.AddItem(policy)).To(Succeed())
			}
			return list
		}

		It("should build inbound rules", func() {
			input, err := os.ReadFile("../matchers/testdata/matchedpolicies/fromrules/01.policies.yaml")
			Expect(err).ToNot(HaveOccurred())

			var policies []model.Resource
			for _, policyBytes := range util_yaml.SplitYAML(string(input)) {
				policy, err := rest.YAML.UnmarshalCore([]byte(policyBytes))
				Expect(err).ToNot(HaveOccurred())
				policies = append(policies, policy)
			}

			listener := core_rules.InboundListener{
				Address: "1.1.1.1",
				Port:    8080,
			}

			actual, err := core_rules.BuildFromRules(map[core_rules.InboundListener]model.ResourceList{
				listener: samePolicyTypesToList(policies),
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual.InboundRules).To(HaveKey(listener))

			bytes, err := yaml.Marshal(actual.InboundRules[listener])
			Expect(err).ToNot(HaveOccurred())
			Expect(string(bytes)).To(ContainSubstring("spiffe://mesh-1/backend"))
			Expect(string(bytes)).To(ContainSubstring("spiffe://mesh-1/orders"))
			Expect(string(bytes)).To(ContainSubstring("spiffe://mesh-1"))
		})

		DescribeTable("should build proxy-wide config for single item policies",
			func(inputFile string) {
				policies := file.ReadInputFile(inputFile)
				rules, err := core_rules.BuildProxyConf(policies)
				Expect(err).ToNot(HaveOccurred())

				bytes, err := yaml.Marshal(rules)
				Expect(err).ToNot(HaveOccurred())
				Expect(bytes).To(matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.yaml", ".golden.yaml", 1)))
			},
			test.EntriesForFolder("rules/single"),
		)
	})
})
