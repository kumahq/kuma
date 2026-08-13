package resolve

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/core"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

type ResourceSection struct {
	Resource    core_model.Resource
	SectionName string
}

func (rs *ResourceSection) Identifier() kri.Identifier {
	return kri.WithSectionName(kri.From(rs.Resource), rs.SectionName)
}

type query struct {
	byLabels    map[string]string
	sectionName string
}

func (q query) findPort(ports []core.Port) core.Port {
	if q.sectionName != "" {
		for _, port := range ports {
			if port.GetName() == q.sectionName {
				return port
			}
		}
	}
	return nil
}

func TargetRef(targetRef common_api.TargetRef, tMeta core_model.ResourceMeta, reader kri.ResourceReader) []*ResourceSection {
	// Removed kinds (e.g. the legacy MeshSubset and MeshServiceSubset) are
	// rejected on create/update but stay readable, so a stored one must not be
	// looked up as a resource type.
	if !targetRef.Kind.IsKnownKind() {
		return nil
	}

	rtype := core_model.ResourceType(targetRef.Kind)

	// Mesh is a real resource but, unlike MeshService and friends, it carries no
	// labels and is selected as the singleton the policy belongs to. Resolve it
	// by identity so that 'to: {kind: Mesh}' keeps producing the mesh-wide
	// ResourceRule that every destination merges in via ResourceRules.Compute.
	if targetRef.Kind == common_api.Mesh {
		if tMeta == nil {
			return []*ResourceSection{}
		}
		if mesh := reader.Get(kri.Identifier{ResourceType: rtype, Name: tMeta.GetMesh()}); mesh != nil {
			return []*ResourceSection{{Resource: mesh}}
		}
		return []*ResourceSection{}
	}

	// targetRef to query
	q := query{
		byLabels:    pointer.Deref(targetRef.Labels),
		sectionName: pointer.Deref(targetRef.SectionName),
	}

	// resolve query without taking port/sectionName into account
	var resources []core_model.Resource
	if q.byLabels != nil {
		list := reader.ListOrEmpty(rtype).GetItems()
		trLabels := subsetutils.NewSubset(q.byLabels)
		for _, r := range list {
			rLabels := subsetutils.NewSubset(r.GetMeta().GetLabels())
			if trLabels.IsSubset(rLabels) {
				resources = append(resources, r)
			}
		}
	}

	if len(resources) == 0 {
		return []*ResourceSection{}
	}

	if q.sectionName == "" {
		result := make([]*ResourceSection, len(resources))
		for i := range resources {
			result[i] = &ResourceSection{Resource: resources[i]}
		}
		return result
	}

	// filter out resources that don't have requested section name or port
	var result []*ResourceSection
	for _, r := range resources {
		if resourceWithPorts, ok := r.(core.Destination); ok {
			if port := q.findPort(resourceWithPorts.GetPorts()); port != nil {
				result = append(result, &ResourceSection{
					Resource:    r,
					SectionName: port.GetName(),
				})
			}
		} else {
			result = append(result, &ResourceSection{Resource: r})
		}
	}

	return result
}
