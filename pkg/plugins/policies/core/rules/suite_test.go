package rules_test

import (
	"strings"
	"testing"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/test"
)

func TestRules(t *testing.T) {
	test.RunSpecs(t, "Rules Suite")
}

func matchedPolicies(rs []core_model.Resource) []core_model.Resource {
	var matched []core_model.Resource
	for _, p := range rs {
		if strings.HasPrefix(p.GetMeta().GetName(), "matched-for-rules-") {
			matched = append(matched, p)
		}
	}
	return matched
}
