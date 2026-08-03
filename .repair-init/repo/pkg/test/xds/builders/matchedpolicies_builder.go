package builders

import (
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
)

type MatchedPoliciesBuilder struct {
	res *xds.MatchedPolicies
}

func MatchedPolicies() *MatchedPoliciesBuilder {
	return &MatchedPoliciesBuilder{res: &xds.MatchedPolicies{
		Dynamic: map[core_model.ResourceType]xds.TypedMatchingPolicies{},
	}}
}

func (mp *MatchedPoliciesBuilder) Build() *xds.MatchedPolicies {
	return mp.res
}

func (mp *MatchedPoliciesBuilder) With(fn func(policies *xds.MatchedPolicies)) *MatchedPoliciesBuilder {
	fn(mp.res)
	return mp
}

func (mp *MatchedPoliciesBuilder) WithToPolicy(resourceType core_model.ResourceType, toRules rules.ToRules) *MatchedPoliciesBuilder {
	mp.res.Dynamic[resourceType] = xds.TypedMatchingPolicies{
		Type:    resourceType,
		ToRules: toRules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithFromPolicy(resourceType core_model.ResourceType, fromRules rules.FromRules) *MatchedPoliciesBuilder {
	mp.res.Dynamic[resourceType] = xds.TypedMatchingPolicies{
		Type:      resourceType,
		FromRules: fromRules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithGatewayPolicy(resourceType core_model.ResourceType, rules rules.GatewayRules) *MatchedPoliciesBuilder {
	mp.res.Dynamic[resourceType] = xds.TypedMatchingPolicies{
		Type:         resourceType,
		GatewayRules: rules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithSingleItemPolicy(resourceType core_model.ResourceType, singleItemRules rules.SingleItemRules) *MatchedPoliciesBuilder {
	mp.res.Dynamic[resourceType] = xds.TypedMatchingPolicies{
		SingleItemRules: singleItemRules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithPolicy(resourceType core_model.ResourceType, toRules rules.ToRules, fromRules rules.FromRules) *MatchedPoliciesBuilder {
	mp.res.Dynamic[resourceType] = xds.TypedMatchingPolicies{
		Type:      resourceType,
		ToRules:   toRules,
		FromRules: fromRules,
	}
	return mp
}
