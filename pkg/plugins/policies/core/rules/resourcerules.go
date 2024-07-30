package rules

import (
	"fmt"
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	meshextenralservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	meshtcproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshtcproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type ResourceRule struct {
	Resource core_model.ResourceMeta
	Conf     interface{}
	Origin   []core_model.ResourceMeta
}

type UniqueResourceKey string

type ResourceRules map[UniqueResourceKey]ResourceRule

func (rr ResourceRules) Compute(r core_model.Resource, l ResourceLister) *ResourceRule {
	key := uniqueKey(r)
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
		coreItem := corePolicyItem(item)
		resolvedItems = append(resolvedItems, resolveTargetRef(coreItem, l)...)
	}

	for _, i := range resolvedItems {
		key := uniqueKey(i.r)
		if _, ok := rules[key]; ok {
			continue
		}

		confs := []interface{}{}
		var origins []core_model.ResourceMeta
		originSet := map[core_model.ResourceKey]struct{}{}
		for _, j := range resolvedItems {
			if includes(j.r, i.r) {
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
				Resource: i.r.GetMeta(),
				Conf:     merged[0],
				Origin:   origins,
			}
		}
	}

	return rules, nil
}

func uniqueKey(r core_model.Resource) UniqueResourceKey {
	switch r.Descriptor().Scope {
	case core_model.ScopeMesh:
		return meshScopedUniqueKey(r.Descriptor().Name, r.GetMeta().GetMesh(), r.GetMeta().GetName())
	default:
		return globalScopedUniqueKey(r.Descriptor().Name, r.GetMeta().GetName())
	}
}

func meshScopedUniqueKey(rtype core_model.ResourceType, mesh, name string) UniqueResourceKey {
	return UniqueResourceKey(fmt.Sprintf("%s.%s.%s", rtype, mesh, name))
}

func globalScopedUniqueKey(rtype core_model.ResourceType, name string) UniqueResourceKey {
	return UniqueResourceKey(fmt.Sprintf("%s.%s", rtype, name))
}

// includes if resource 'y' is part of the resource 'x', i.e. 'MeshService' is always included in 'Mesh'
func includes(x, y core_model.Resource) bool {
	switch x.Descriptor().Name {
	case core_mesh.MeshType:
		switch y.Descriptor().Name {
		case core_mesh.MeshType:
			return x.GetMeta().GetName() == y.GetMeta().GetName()
		default:
			return x.GetMeta().GetName() == y.GetMeta().GetMesh()
		}
	case meshservice_api.MeshServiceType:
		switch y.Descriptor().Name {
		case meshservice_api.MeshServiceType:
			return uniqueKey(x) == uniqueKey(y)
		default:
			return false
		}
	case meshextenralservice_api.MeshExternalServiceType:
		switch y.Descriptor().Name {
		case meshextenralservice_api.MeshExternalServiceType:
			return uniqueKey(x) == uniqueKey(y)
		default:
			return false
		}
	default:
		return false
	}
}

type resolvedPolicyItem struct {
	item PolicyItemWithMeta
	r    core_model.Resource
}

func resolveTargetRef(item PolicyItemWithMeta, l ResourceLister) []*resolvedPolicyItem {
	if !item.GetTargetRef().Kind.IsRealResource() {
		return nil
	}

	list := l.ListOrEmpty(core_model.ResourceType(item.GetTargetRef().Kind)).GetItems()

	switch {
	case item.GetTargetRef().Name != "":
		searchName := item.GetTargetRef().Name
		if i := slices.IndexFunc(list, func(r core_model.Resource) bool { return r.GetMeta().GetName() == searchName }); i >= 0 {
			return []*resolvedPolicyItem{{r: list[i], item: item}}
		}
	case len(item.GetTargetRef().Labels) > 0:
		var rv []*resolvedPolicyItem
		trLabels := NewSubset(item.GetTargetRef().Labels)
		for _, r := range list {
			rLabels := NewSubset(r.GetMeta().GetLabels())
			if trLabels.IsSubset(rLabels) {
				rv = append(rv, &resolvedPolicyItem{r: r, item: item})
			}
		}
		return rv
	}

	return nil
}

func corePolicyItem(item PolicyItemWithMeta) PolicyItemWithMeta {
	policyConf := item.PolicyItem.GetDefault()

	switch conf := policyConf.(type) {
	case meshhttproute_api.PolicyDefault:
		for i, rule := range conf.Rules {
			conf.Rules[i].Default.BackendRefs = pointer.To(coreBackendRefs(item.ResourceMeta, pointer.Deref(rule.Default.BackendRefs)))
		}
		policyConf = conf
	case meshtcproute_api.Rule:
		conf.Default.BackendRefs = coreBackendRefs(item.ResourceMeta, conf.Default.BackendRefs)
		policyConf = conf
	}

	switch item.PolicyItem.GetTargetRef().Kind {
	case common_api.Mesh:
		return PolicyItemWithMeta{
			ResourceMeta: item.ResourceMeta,
			PolicyItem: &artificialPolicyItem{
				conf: policyConf,
				targetRef: common_api.TargetRef{
					Kind: common_api.Mesh,
					Name: item.GetMesh(),
				},
			},
		}
	case common_api.MeshService, common_api.MeshExternalService:
		return PolicyItemWithMeta{
			ResourceMeta: item.ResourceMeta,
			PolicyItem: &artificialPolicyItem{
				conf:      policyConf,
				targetRef: core_model.CoreTargetRef(item.ResourceMeta, item.PolicyItem.GetTargetRef()),
			},
		}
	default:
		return item
	}
}

func coreBackendRefs(rm core_model.ResourceMeta, backendRefs []common_api.BackendRef) []common_api.BackendRef {
	var rv []common_api.BackendRef
	for _, br := range backendRefs {
		rv = append(rv, common_api.BackendRef{
			TargetRef: core_model.CoreTargetRef(rm, br.TargetRef),
			Weight:    br.Weight,
		})
	}
	return rv
}
