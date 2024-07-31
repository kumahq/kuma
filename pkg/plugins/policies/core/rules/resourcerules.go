package rules

import (
	"fmt"
	"slices"
	"strings"

	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshextenralservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
)

type ResourceRule struct {
	Resource core_model.ResourceMeta
	Conf     []interface{}
	Origin   []core_model.ResourceMeta
}

type UniqueResourceIdentifier struct {
	core_model.ResourceIdentifier

	ResourceType core_model.ResourceType
	SectionName  string
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

type UniqueResourceKey string

type ResourceRules map[UniqueResourceKey]ResourceRule

type ComputeOpts struct {
	sectionName string
}

func NewComputeOpts(fn ...ComputeOptsFn) *ComputeOpts {
	opts := &ComputeOpts{}
	for _, f := range fn {
		f(opts)
	}
	return opts
}

type ComputeOptsFn func(*ComputeOpts)

func WithSectionName(sectionName string) ComputeOptsFn {
	return func(opts *ComputeOpts) {
		opts.sectionName = sectionName
	}
}

func (rr ResourceRules) Compute(r core_model.Resource, l ResourceLister, fn ...ComputeOptsFn) *ResourceRule {
	opts := NewComputeOpts(fn...)
	key := uniqueKey(r, opts.sectionName)
	if rule, ok := rr[UniqueResourceKey(key.String())]; ok {
		return &rule
	}

	findMesh := func(meshName string) core_model.Resource {
		meshes := l.ListOrEmpty(core_mesh.MeshType).GetItems()
		if i := slices.IndexFunc(meshes, func(resource core_model.Resource) bool {
			return resource.GetMeta().GetName() == meshName
		}); i >= 0 {
			return meshes[i]
		}
		return nil
	}

	switch r.Descriptor().Name {
	case meshservice_api.MeshServiceType:
		// find MeshService without the sectionName and compute rules for it
		if opts.sectionName != "" {
			return rr.Compute(r, l)
		}
		// find MeshService's Mesh and compute rules for it
		if mesh := findMesh(r.GetMeta().GetMesh()); mesh != nil {
			return rr.Compute(mesh, l)
		}
	case meshextenralservice_api.MeshExternalServiceType:
		// find MeshExternalService's Mesh and compute rules for it
		if mesh := findMesh(r.GetMeta().GetMesh()); mesh != nil {
			return rr.Compute(mesh, l)
		}
	case meshhttproute_api.MeshHTTPRouteType:
		// todo(lobkovilya): handle MeshHTTPRoute
	}

	return nil
}

func BuildResourceRules(list []PolicyItemWithMeta, l ResourceLister) (ResourceRules, error) {
	rules := ResourceRules{}

	SortByTargetRefV2(list)

	var resolvedItems []*resolvedPolicyItem
	for _, item := range list {
		resolvedItems = append(resolvedItems, resolveTargetRef(item, l)...)
	}

	// we could've built ResourceRule for all resources in the cluster, but we only need to build rules for resources
	// that are part of the policy to reduce the size of the ResourceRules
	for uri, resource := range indexResources(resolvedItems) {
		// take only policy items that have isRelevant conf for the resource
		var relevant []*resolvedPolicyItem
		for _, policyItem := range resolvedItems {
			if isRelevant(policyItem, resource, uri.SectionName) {
				relevant = append(relevant, policyItem)
			}
		}

		if len(relevant) > 0 {
			// merge all relevant confs into one, the order of merging is guaranteed by SortByTargetRefV2
			merged, err := mergeConfs(relevant)
			if err != nil {
				return nil, err
			}
			rules[UniqueResourceKey(uri.String())] = ResourceRule{
				Resource: resource.GetMeta(),
				Conf:     merged,
				Origin:   origins(relevant),
			}
		}

		//// merge all relevant confs into one, the order of merging is guaranteed by SortByTargetRefV2
		//merged, err := mergeConfs(relevant)
		//if err != nil {
		//	return nil, err
		//}
		//
		//if len(merged) == 1 {
		//	rules[UniqueResourceKey(uri.String())] = ResourceRule{
		//		Resource: resource.GetMeta(),
		//		Conf:     merged[0],
		//		Origin:   origins(relevant),
		//	}
		//}
	}

	return rules, nil
}

func indexResources(ri []*resolvedPolicyItem) map[UniqueResourceIdentifier]core_model.Resource {
	index := map[UniqueResourceIdentifier]core_model.Resource{}
	for _, i := range ri {
		index[uniqueKey(i.resource, i.sectionName())] = i.resource
	}
	return index
}

func mergeConfs(ri []*resolvedPolicyItem) ([]interface{}, error) {
	var confs []interface{}
	for _, i := range ri {
		confs = append(confs, i.item.GetDefault())
	}
	return MergeConfs(confs)
}

func origins(ri []*resolvedPolicyItem) []core_model.ResourceMeta {
	var rv []core_model.ResourceMeta
	set := map[core_model.ResourceKey]struct{}{}
	for _, i := range ri {
		if _, ok := set[core_model.MetaToResourceKey(i.item.ResourceMeta)]; !ok {
			rv = append(rv, i.item.ResourceMeta)
			set[core_model.MetaToResourceKey(i.item.ResourceMeta)] = struct{}{}
		}
	}
	return rv
}

func uniqueKey(r core_model.Resource, sectionName string) UniqueResourceIdentifier {
	return UniqueResourceIdentifier{
		ResourceIdentifier: core_model.NewResourceIdentifier(r),
		ResourceType:       r.Descriptor().Name,
		SectionName:        sectionName,
	}
}

// isRelevant returns true if the policyItem is relevant to the resource or section of the resource
func isRelevant(policyItem *resolvedPolicyItem, r core_model.Resource, sectionName string) bool {
	switch policyItem.resource.Descriptor().Name {
	case core_mesh.MeshType:
		switch r.Descriptor().Name {
		case core_mesh.MeshType:
			return policyItem.resource.GetMeta().GetName() == r.GetMeta().GetName()
		default:
			return policyItem.resource.GetMeta().GetName() == r.GetMeta().GetMesh()
		}
	case meshservice_api.MeshServiceType:
		switch r.Descriptor().Name {
		case meshservice_api.MeshServiceType:
			switch {
			case uniqueKey(policyItem.resource, policyItem.sectionName()) == uniqueKey(r, sectionName):
				return true
			case uniqueKey(policyItem.resource, "") == uniqueKey(r, "") && policyItem.sectionName() == "" && sectionName != "":
				return true
			default:
				return false
			}
		default:
			return false
		}
	case meshextenralservice_api.MeshExternalServiceType:
		switch r.Descriptor().Name {
		case meshextenralservice_api.MeshExternalServiceType:
			return uniqueKey(policyItem.resource, "") == uniqueKey(r, "")
		default:
			return false
		}
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

func resolveTargetRef(item PolicyItemWithMeta, l ResourceLister) []*resolvedPolicyItem {
	if !item.GetTargetRef().Kind.IsRealResource() {
		return nil
	}

	list := l.ListOrEmpty(core_model.ResourceType(item.GetTargetRef().Kind)).GetItems()

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
	if i := slices.IndexFunc(list, func(r core_model.Resource) bool {
		return ri == core_model.NewResourceIdentifier(r)
	}); i >= 0 {
		return []*resolvedPolicyItem{{resource: list[i], item: item}}
	}

	return nil
}
