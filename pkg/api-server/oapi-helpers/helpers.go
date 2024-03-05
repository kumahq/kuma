package oapi_helpers

import (
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
)

func ResourceToMeta(m core_model.Resource) api_common.Meta {
	return ResourceMetaToMeta(m.Descriptor().Name, m.GetMeta())
}

func ResourceMetaToMeta(resType core_model.ResourceType, m core_model.ResourceMeta) api_common.Meta {
	return api_common.Meta{
		Type: string(resType),
		Mesh: m.GetMesh(),
		Name: m.GetName(),
	}
}

func ResourceMetaListToMetaList(resType core_model.ResourceType, in []core_model.ResourceMeta) []api_common.Meta {
	out := make([]api_common.Meta, len(in))
	for i, o := range in {
		out[i] = ResourceMetaToMeta(resType, o)
	}
	return out
}

func SubsetToRuleMatcher(subset core_rules.Subset) []api_common.RuleMatcher {
	matchers := []api_common.RuleMatcher{}
	for _, m := range subset {
		matchers = append(matchers, api_common.RuleMatcher{Key: m.Key, Value: m.Value, Not: m.Not})
	}
	return matchers
}
