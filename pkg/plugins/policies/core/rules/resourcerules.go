package rules

import (
	"fmt"
	"strings"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshexternalservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshmultizoneservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshmultizoneservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
)

type ResourceRule struct {
	Resource            core_model.ResourceMeta
	ResourceSectionName string
	Conf                []interface{}
	Origin              []Origin

	// BackendRefOriginIndex is a mapping from the rule to the origin of the BackendRefs in the rule.
	// Some policies have BackendRefs in their confs, and it's important to know what was the original policy
	// that contributed the BackendRefs to the final conf. Rule (key) is represented as a hash from rule.Matches.
	// Origin (value) is represented as an index in the Origin list. If policy doesn't have rules (i.e. MeshTCPRoute)
	// then key is an empty string "".
	BackendRefOriginIndex BackendRefOriginIndex
}

type BackendRefOriginIndex map[MatchesHash]int

func (originIndex BackendRefOriginIndex) Update(conf interface{}, newIndex int) {
	switch conf := conf.(type) {
	case meshtcproute_api.Rule:
		if conf.Default.BackendRefs != nil {
			originIndex[EmptyMatches] = newIndex
		}
	case meshhttproute_api.PolicyDefault:
		for _, rule := range conf.Rules {
			if rule.Default.BackendRefs != nil {
				hash := meshhttproute_api.HashMatches(rule.Matches)
				originIndex[MatchesHash(hash)] = newIndex
			}
		}
	default:
		return
	}
}

type MatchesHash string

var EmptyMatches MatchesHash = ""

type Origin struct {
	Resource core_model.ResourceMeta
	// RuleIndex is an index in the 'to[]' array, so we could unambiguously detect what to-item contributed to the final conf.
	// Especially useful when to-item uses `targetRef.Labels`, because there is no obvious matching between the specific resource
	// in `ResourceRule.Resource` and to-item.
	RuleIndex int
}

type UniqueResourceIdentifier struct {
	core_model.ResourceIdentifier

	ResourceType core_model.ResourceType
	SectionName  string
}

func (ri UniqueResourceIdentifier) MarshalText() ([]byte, error) {
	return []byte(ri.String()), nil
}

func (ri UniqueResourceIdentifier) String() string {
	var pairs []string
	if ri.ResourceType != "" {
		pairs = append(pairs, strings.ToLower(string(ri.ResourceType)))
	}
	pairs = append(pairs, ri.ResourceIdentifier.String())
	if ri.SectionName != "" {
		pairs = append(pairs, fmt.Sprintf("section/%s", ri.SectionName))
	}
	return strings.Join(pairs, ":")
}

type ResourceRules map[UniqueResourceIdentifier]ResourceRule

func (rr ResourceRules) Compute(uri UniqueResourceIdentifier, reader ResourceReader) *ResourceRule {
	if rule, ok := rr[uri]; ok {
		return &rule
	}

	switch uri.ResourceType {
	case meshservice_api.MeshServiceType, meshmultizoneservice_api.MeshMultiZoneServiceType:
		// find MeshService without the sectionName and compute rules for it
		if uri.SectionName != "" {
			uriWithoutSection := uri
			uriWithoutSection.SectionName = ""
			return rr.Compute(uriWithoutSection, reader)
		}
		// find MeshService's Mesh and compute rules for it
		if mesh := reader.Get(core_mesh.MeshType, core_model.ResourceIdentifier{Name: uri.Mesh}); mesh != nil {
			return rr.Compute(UniqueKey(mesh, ""), reader)
		}
	case meshexternalservice_api.MeshExternalServiceType:
		// find MeshExternalService's Mesh and compute rules for it
		if mesh := reader.Get(core_mesh.MeshType, core_model.ResourceIdentifier{Name: uri.Mesh}); mesh != nil {
			return rr.Compute(UniqueKey(mesh, ""), reader)
		}
	case meshhttproute_api.MeshHTTPRouteType:
		// todo(lobkovilya): handle MeshHTTPRoute
	}

	return nil
}

func BuildResourceRules(list []PolicyItemWithMeta, reader ResourceReader) (ResourceRules, error) {
	rules := ResourceRules{}

	SortByTargetRefV2(list)

	var resolvedItems []*resolvedPolicyItem
	for _, item := range list {
		resolvedItems = append(resolvedItems, resolveTargetRef(item, reader)...)
	}

	// we could've built ResourceRule for all resources in the cluster, but we only need to build rules for resources
	// that are part of the policy to reduce the size of the ResourceRules
	for uri, resource := range indexResources(resolvedItems) {
		// take only policy items that have isRelevant conf for the resource
		var relevant []PolicyItemWithMeta
		for _, policyItem := range resolvedItems {
			if isRelevant(policyItem, resource, uri.SectionName) {
				relevant = append(relevant, policyItem.item)
			}
		}

		if len(relevant) > 0 {
			// merge all relevant confs into one, the order of merging is guaranteed by SortByTargetRefV2
			merged, err := mergeConfs(relevant)
			if err != nil {
				return nil, err
			}
			ruleOrigins, originIndex := origins(relevant, true)
			rules[uri] = ResourceRule{
				Resource:              resource.GetMeta(),
				ResourceSectionName:   uri.SectionName,
				Conf:                  merged,
				Origin:                ruleOrigins,
				BackendRefOriginIndex: originIndex,
			}
		}
	}

	return rules, nil
}

func indexResources(ri []*resolvedPolicyItem) map[UniqueResourceIdentifier]core_model.Resource {
	index := map[UniqueResourceIdentifier]core_model.Resource{}
	for _, i := range ri {
		index[UniqueKey(i.resource, i.sectionName())] = i.resource
	}
	return index
}

func mergeConfs(items []PolicyItemWithMeta) ([]interface{}, error) {
	var confs []interface{}
	for _, item := range items {
		confs = append(confs, item.GetDefault())
	}
	return MergeConfs(confs)
}

func origins(items []PolicyItemWithMeta, withRuleIndex bool) ([]Origin, BackendRefOriginIndex) {
	var rv []Origin

	type keyType struct {
		core_model.ResourceKey
		ruleIndex int
	}
	key := func(policyItem PolicyItemWithMeta) keyType {
		k := keyType{
			ResourceKey: core_model.MetaToResourceKey(policyItem.ResourceMeta),
		}
		if withRuleIndex {
			k.ruleIndex = policyItem.RuleIndex
		}
		return k
	}
	set := map[keyType]struct{}{}
	originIndex := BackendRefOriginIndex{}
	for _, item := range items {
		if _, ok := set[key(item)]; !ok {
			originIndex.Update(item.GetDefault(), len(rv))
			rv = append(rv, Origin{Resource: item.ResourceMeta, RuleIndex: item.RuleIndex})
			set[key(item)] = struct{}{}
		}
	}
	return rv, originIndex
}

func UniqueKey(r core_model.Resource, sectionName string) UniqueResourceIdentifier {
	return UniqueResourceIdentifier{
		ResourceIdentifier: core_model.NewResourceIdentifier(r),
		ResourceType:       r.Descriptor().Name,
		SectionName:        sectionName,
	}
}

// isRelevant returns true if the policyItem is relevant to the resource or section of the resource
func isRelevant(policyItem *resolvedPolicyItem, r core_model.Resource, sectionName string) bool {
	switch itemDescriptorName := policyItem.resource.Descriptor().Name; itemDescriptorName {
	case core_mesh.MeshType:
		switch r.Descriptor().Name {
		case core_mesh.MeshType:
			return policyItem.resource.GetMeta().GetName() == r.GetMeta().GetName()
		default:
			return policyItem.resource.GetMeta().GetName() == r.GetMeta().GetMesh()
		}
	case meshservice_api.MeshServiceType, meshmultizoneservice_api.MeshMultiZoneServiceType:
		if r.Descriptor().Name != itemDescriptorName {
			return false
		}
		switch {
		case UniqueKey(policyItem.resource, policyItem.sectionName()) == UniqueKey(r, sectionName):
			return true
		case UniqueKey(policyItem.resource, "") == UniqueKey(r, "") && policyItem.sectionName() == "" && sectionName != "":
			return true
		default:
			return false
		}
	case meshexternalservice_api.MeshExternalServiceType:
		if r.Descriptor().Name != itemDescriptorName {
			return false
		}
		return UniqueKey(policyItem.resource, "") == UniqueKey(r, "")
	default:
		return false
	}
}

type resolvedPolicyItem struct {
	item     PolicyItemWithMeta
	resource core_model.Resource
}

func (r *resolvedPolicyItem) sectionName() string {
	return r.item.PolicyItem.GetTargetRef().SectionName
}

func resolveTargetRef(item PolicyItemWithMeta, reader ResourceReader) []*resolvedPolicyItem {
	if !item.GetTargetRef().Kind.IsRealResource() {
		return nil
	}
	rtype := core_model.ResourceType(item.GetTargetRef().Kind)
	list := reader.ListOrEmpty(rtype).GetItems()

	if len(item.GetTargetRef().Labels) > 0 {
		var rv []*resolvedPolicyItem
		trLabels := NewSubset(item.GetTargetRef().Labels)
		for _, r := range list {
			rLabels := NewSubset(r.GetMeta().GetLabels())
			if trLabels.IsSubset(rLabels) {
				rv = append(rv, &resolvedPolicyItem{resource: r, item: item})
			}
		}
		return rv
	}

	ri := core_model.TargetRefToResourceIdentifier(item.ResourceMeta, item.GetTargetRef())
	if resource := reader.Get(rtype, ri); resource != nil {
		return []*resolvedPolicyItem{{resource: resource, item: item}}
	}

	return nil
}
