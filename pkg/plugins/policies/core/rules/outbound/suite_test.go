package outbound_test

import (
	"strings"
	"testing"

	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test"
)

func TestRules(t *testing.T) {
	test.RunSpecs(t, "Outbound Rules Suite")
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
