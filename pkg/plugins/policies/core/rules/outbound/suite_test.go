package outbound_test

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/test"
)

func TestRules(t *testing.T) {
	test.RunSpecs(t, "Outbound Rules Suite")
}

func matchedPolicies(rs []core_model.Resource) core_model.ResourceList {
	Expect(rs).ToNot(BeEmpty())

	var matchedRs []core_model.Resource
	for _, p := range rs {
		if strings.HasPrefix(p.GetMeta().GetName(), "matched-for-rules-") {
			matchedRs = append(matchedRs, p)
		}
	}

	Expect(matchedRs).ToNot(BeEmpty())
	policyType := matchedRs[0].Descriptor().Name

	list, err := registry.Global().NewList(policyType)
	Expect(err).ToNot(HaveOccurred())

	for _, p := range matchedRs {
		Expect(list.AddItem(p)).To(Succeed())
	}
	return list
}
