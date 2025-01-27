package inbound_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/file"
)

var _ = Describe("BuildInboundRules", func() {
	buildResourceList := func(rs []core_model.Resource) core_model.ResourceList {
		Expect(rs).ToNot(BeEmpty())
		rl := rs[0].Descriptor().NewList()
		for _, p := range rs {
			if strings.HasPrefix(p.GetMeta().GetName(), "matched-for-rules-") {
				_ = rl.AddItem(p)
			}
		}
		return rl
	}

	DescribeTable("should build a rule-based view for policies",
		func(inputFile string) {
			// given
			resources := file.ReadInputFile(inputFile)
			resourceList := buildResourceList(resources)

			// when
			rules, err := inbound.BuildRules(resourceList)
			Expect(err).ToNot(HaveOccurred())

			// then
			bytes, err := yaml.Marshal(struct {
				Rules []*inbound.Rule `json:"rules"`
			}{Rules: rules})
			Expect(err).ToNot(HaveOccurred())
			Expect(bytes).To(matchers.MatchGoldenYAML(strings.Replace(inputFile, ".input.", ".golden.", 1)))
		},
		test.EntriesForFolder("inboundrules"),
	)
})
