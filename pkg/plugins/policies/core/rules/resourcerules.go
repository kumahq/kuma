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
	Conf     interface{}
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
	if rule, ok := rr[key]; ok {
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

	for _, i := range resolvedItems {
		key := uniqueKey(i.resource, i.sectionName())
		if _, ok := rules[key]; ok {
			continue
		}

		confs := []interface{}{}
		var origins []core_model.ResourceMeta
		originSet := map[core_model.ResourceKey]struct{}{}
		for _, j := range resolvedItems {
			if includes(j, i) {
				confs = append(confs, j.item.GetDefault())
				if _, ok := originSet[core_model.MetaToResourceKey(j.item.ResourceMeta)]; !ok {
					origins = append(origins, j.item.ResourceMeta)
					originSet[core_model.MetaToResourceKey(j.item.ResourceMeta)] = struct{}{}
				}
			}
		}

		merged, err := MergeConfs(confs)
		if err != nil {
			return nil, err
		}
		if len(merged) == 1 {
			rules[key] = ResourceRule{
				Resource: i.resource.GetMeta(),
				Conf:     merged[0],
				Origin:   origins,
			}
		}
	}

	return rules, nil
}

func uniqueKey(r core_model.Resource, sectionName string) UniqueResourceKey {
	tri := UniqueResourceIdentifier{
		ResourceIdentifier: core_model.NewResourceIdentifier(r),
		ResourceType:       r.Descriptor().Name,
		SectionName:        sectionName,
	}

	return UniqueResourceKey(tri.String())
}

// includes if resource 'y' is part of the resource 'x', i.e. 'MeshService' is always included in 'Mesh'
func includes(x, y *resolvedPolicyItem) bool {
	switch x.resource.Descriptor().Name {
	case core_mesh.MeshType:
		switch y.resource.Descriptor().Name {
		case core_mesh.MeshType:
			return x.resource.GetMeta().GetName() == y.resource.GetMeta().GetName()
		default:
			return x.resource.GetMeta().GetName() == y.resource.GetMeta().GetMesh()
		}
	case meshservice_api.MeshServiceType:
		switch y.resource.Descriptor().Name {
		case meshservice_api.MeshServiceType:
			switch {
			case uniqueKey(x.resource, x.sectionName()) == uniqueKey(y.resource, y.sectionName()):
				return true
			case uniqueKey(x.resource, "") == uniqueKey(y.resource, "") && x.sectionName() == "" && y.sectionName() != "":
				return true
			default:
				return false
			}
		default:
			return false
		}
	case meshextenralservice_api.MeshExternalServiceType:
		switch y.resource.Descriptor().Name {
		case meshextenralservice_api.MeshExternalServiceType:
			return uniqueKey(x.resource, "") == uniqueKey(y.resource, "")
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
