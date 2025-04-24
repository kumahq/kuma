package policies

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/outbound"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

func newTestResourceMeta() *test_model.ResourceMeta {
	return &test_model.ResourceMeta{Mesh: "default", Name: "test-origin"}
}

func newTestOrigin() common.Origin {
	return common.Origin{
		Resource:  newTestResourceMeta(),
		RuleIndex: 0,
	}
}

func NewRule(s subsetutils.Subset, conf interface{}) *core_rules.Rule {
	originByMatches := map[common_api.MatchesHash]core_model.ResourceMeta{}

	switch c := conf.(type) {
	case meshtcproute_api.Rule:
		originByMatches[common.EmptyMatches] = newTestResourceMeta()
	case meshhttproute_api.PolicyDefault:
		for _, rule := range c.Rules {
			originByMatches[meshhttproute_api.HashMatches(rule.Matches)] = newTestResourceMeta()
		}
	}

	return &core_rules.Rule{
		Subset:          s,
		Conf:            conf,
		Origin:          []core_model.ResourceMeta{newTestResourceMeta()},
		OriginByMatches: originByMatches,
	}
}

func NewOutboundRule(r core_model.ResourceMeta, conf interface{}) outbound.ResourceRule {
	originByMatches := map[common_api.MatchesHash]common.Origin{}

	switch c := conf.(type) {
	case meshtcproute_api.Rule:
		originByMatches[common.EmptyMatches] = newTestOrigin()
	case meshhttproute_api.PolicyDefault:
		for _, rule := range c.Rules {
			originByMatches[meshhttproute_api.HashMatches(rule.Matches)] = newTestOrigin()
		}
	}

	return outbound.ResourceRule{
		Resource:        r,
		Conf:            []interface{}{conf},
		Origin:          []common.Origin{newTestOrigin()},
		OriginByMatches: originByMatches,
	}
}
