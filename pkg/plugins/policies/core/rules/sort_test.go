package rules_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/file"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
)

var _ = Describe("SortByTargetRefV2", func() {
	DescribeTable("should sort to-items",
		func(inputFile string) {
			// given
			resources := file.ReadInputFile(inputFile)
			meshCtx := xds_builders.Context().WithMeshLocalResources(resources).Build()
			toList, err := core_rules.BuildToList(matchedPolicies(resources), meshCtx.Mesh.Resources)
			Expect(err).ToNot(HaveOccurred())

			// when
			core_rules.SortByTargetRefV2(toList)

			// then
			bytes, err := yaml.Marshal(toList)
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.", ".golden.", 1)))
		},
		test.EntriesForFolder("sort"),
	)
})
