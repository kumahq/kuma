package builders

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
)

type MatchedPoliciesBuilder struct {
	plugins xds.PluginOriginatedPolicies
}

func MatchedPolicies() *MatchedPoliciesBuilder {
	return &MatchedPoliciesBuilder{
		plugins: xds.PluginOriginatedPolicies{},
	}
}

func (mp *MatchedPoliciesBuilder) Build() xds.PluginOriginatedPolicies {
	return mp.plugins
}

func (mp *MatchedPoliciesBuilder) WithToPolicy(resourceType core_model.ResourceType, toRules rules.ToRules) *MatchedPoliciesBuilder {
	mp.plugins[resourceType] = xds.TypedMatchingPolicies{
		Type:    resourceType,
		ToRules: toRules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithFromPolicy(resourceType core_model.ResourceType, fromRules rules.FromRules) *MatchedPoliciesBuilder {
	mp.plugins[resourceType] = xds.TypedMatchingPolicies{
		Type:      resourceType,
		FromRules: fromRules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithGatewayPolicy(resourceType core_model.ResourceType, rules rules.GatewayRules) *MatchedPoliciesBuilder {
	mp.plugins[resourceType] = xds.TypedMatchingPolicies{
		Type:         resourceType,
		GatewayRules: rules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithSingleItemPolicy(resourceType core_model.ResourceType, singleItemRules rules.SingleItemRules) *MatchedPoliciesBuilder {
	mp.plugins[resourceType] = xds.TypedMatchingPolicies{
		SingleItemRules: singleItemRules,
	}
	return mp
}

func (mp *MatchedPoliciesBuilder) WithPolicy(resourceType core_model.ResourceType, toRules rules.ToRules, fromRules rules.FromRules) *MatchedPoliciesBuilder {
	mp.plugins[resourceType] = xds.TypedMatchingPolicies{
		Type:      resourceType,
		ToRules:   toRules,
		FromRules: fromRules,
	}
	return mp
}
