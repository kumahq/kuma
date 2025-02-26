package outbound_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/file"
)

var _ = Describe("SortToEntries", func() {
	DescribeTable("should sort to-items",
		func(inputFile string) {
			// given
			resources := file.ReadInputFile(inputFile)
			entries, err := outbound.GetEntries(matchedPolicies(resources))
			Expect(err).ToNot(HaveOccurred())

			// when
			outbound.Sort(entries)

			// then
			bytes, err := yaml.Marshal(entries)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.", ".golden.", 1)))
		},
		test.EntriesForFolder("sort"),
	)
})
