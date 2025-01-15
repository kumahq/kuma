package inbound_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/file"
)

var _ = Describe("BuildInboundRules", func() {
	DescribeTable("should build a rule-based view for policies",
		func(inputFile string) {
			// given
			resources := file.ReadInputFile(inputFile)

			// when
			rules, err := inbound.BuildRules(matchedPolicies(resources))
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(rules)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.", ".golden.", 1)))
		},
		test.EntriesForFolder("inboundrules"),
	)
})
