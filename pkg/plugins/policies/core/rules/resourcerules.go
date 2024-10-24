package rules

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
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

func (r *ResourceRule) GetBackendRefOrigin(hash common_api.MatchesHash) (core_model.ResourceMeta, bool) {
	if r == nil {
		return nil, false
	}
	if r.BackendRefOriginIndex == nil {
		return nil, false
	}
	index, ok := r.BackendRefOriginIndex[hash]
	if !ok {
		return nil, false
	}
	if index >= len(r.Origin) {
		return nil, false
	}
	return r.Origin[index].Resource, true
}

type BackendRefOriginIndex map[common_api.MatchesHash]int

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
				originIndex[hash] = newIndex
			}
		}
	default:
		return
	}
}

var EmptyMatches common_api.MatchesHash = ""

type Origin struct {
	Resource core_model.ResourceMeta
	// RuleIndex is an index in the 'to[]' array, so we could unambiguously detect what to-item contributed to the final conf.
	// Especially useful when to-item uses `targetRef.Labels`, because there is no obvious matching between the specific resource
	// in `ResourceRule.Resource` and to-item.
	RuleIndex int
}

type ResourceRules map[core_model.TypedResourceIdentifier]ResourceRule

func (rr ResourceRules) Compute(uri core_model.TypedResourceIdentifier, reader ResourceReader) *ResourceRule {
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

func indexResources(ri []*resolvedPolicyItem) map[core_model.TypedResourceIdentifier]core_model.Resource {
	index := map[core_model.TypedResourceIdentifier]core_model.Resource{}
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

func UniqueKey(r core_model.Resource, sectionName string) core_model.TypedResourceIdentifier {
	return core_model.TypedResourceIdentifier{
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
	item                PolicyItemWithMeta
	resource            core_model.Resource
	implicitSectionName string
}

func (r *resolvedPolicyItem) sectionName() string {
	if refSectionName := r.item.PolicyItem.GetTargetRef().SectionName; refSectionName != "" {
		return refSectionName
	}
	return r.implicitSectionName
}

func resolveTargetRef(item PolicyItemWithMeta, reader ResourceReader) []*resolvedPolicyItem {
	if !item.GetTargetRef().Kind.IsRealResource() {
		return nil
	}
	rtype := core_model.ResourceType(item.GetTargetRef().Kind)
	list := reader.ListOrEmpty(rtype).GetItems()

	var implicitPort uint32
	implicitLabels := map[string]string{}
	if item.GetTargetRef().Kind == common_api.MeshService && item.GetTargetRef().SectionName == "" {
		if name, namespace, port, err := parseService(item.GetTargetRef().Name); err == nil {
			implicitLabels[mesh_proto.KubeNamespaceTag] = namespace
			implicitLabels[mesh_proto.DisplayName] = name
			implicitPort = port
		}
	}

	labels := item.GetTargetRef().Labels
	if len(implicitLabels) > 0 {
		labels = implicitLabels
	}

	if len(labels) > 0 {
		var rv []*resolvedPolicyItem
		trLabels := NewSubset(labels)
		for _, r := range list {
			rLabels := NewSubset(r.GetMeta().GetLabels())
			var implicitSectionName string
			if ms, ok := r.(*meshservice_api.MeshServiceResource); ok && implicitPort != 0 {
				for _, port := range ms.Spec.Ports {
					if port.Port == implicitPort {
						implicitSectionName = port.Name
					}
				}
			}
			if trLabels.IsSubset(rLabels) {
				rv = append(rv, &resolvedPolicyItem{resource: r, item: item, implicitSectionName: implicitSectionName})
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

func parseService(host string) (string, string, uint32, error) {
	// split host into <name>_<namespace>_svc_<port>
	segments := strings.Split(host, "_")

	var port uint32
	switch len(segments) {
	case 4:
		p, err := strconv.ParseInt(segments[3], 10, 32)
		if err != nil {
			return "", "", 0, err
		}
		port = uint32(p)
	case 3:
		// service less service names have no port, so we just put the reserved
		// one here to note that this service is actually
		port = mesh_proto.TCPPortReserved
	default:
		return "", "", 0, errors.Errorf("service tag in unexpected format")
	}

	name, namespace := segments[0], segments[1]
	return name, namespace, port, nil
}
