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

func TargetRef(targetRef common_api.TargetRef, _ core_model.ResourceMeta, reader kri.ResourceReader) []*ResourceSection {
	if !targetRef.Kind.IsRealResource() {
		return nil
	}

	// targetRef to query
	q := query{
		byLabels:    pointer.Deref(targetRef.Labels),
		sectionName: pointer.Deref(targetRef.SectionName),
	}

	// resolve query without taking port/sectionName into account
	rtype := core_model.ResourceType(targetRef.Kind)
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
