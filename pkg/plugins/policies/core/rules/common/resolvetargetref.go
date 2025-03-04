package common

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/core"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type ResourceSection struct {
	Resource    core_model.Resource
	SectionName string
}

func (rs *ResourceSection) Identifier() core_model.TypedResourceIdentifier {
	return UniqueKey(rs.Resource, rs.SectionName)
}

func UniqueKey(r core_model.Resource, sectionName string) core_model.TypedResourceIdentifier {
	return core_model.TypedResourceIdentifier{
		ResourceIdentifier: core_model.NewResourceIdentifier(r),
		ResourceType:       r.Descriptor().Name,
		SectionName:        sectionName,
	}
}

type ResourceWithPorts interface {
	GetPorts() []core.Port
}

type query struct {
	byIdentifier *core_model.ResourceIdentifier
	byLabels     map[string]string
	port         uint32
	sectionName  string
}

func (q query) findPort(ports []core.Port) core.Port {
	switch {
	case q.port != 0:
		for _, port := range ports {
			if port.GetValue() == q.port {
				return port
			}
		}
	case q.sectionName != "":
		for _, port := range ports {
			if port.GetName() == q.sectionName {
				return port
			}
			if parsed, ok := tryParsePort(q.sectionName); ok && port.GetName() == "" && port.GetValue() == parsed {
				return port
			}
		}
	}
	return nil
}

func ResolveTargetRef(targetRef common_api.TargetRef, tMeta core_model.ResourceMeta, reader ResourceReader) []*ResourceSection {
	if !targetRef.Kind.IsRealResource() {
		return nil
	}

	// targetRef to query
	var q query
	switch {
	case len(pointer.Deref(targetRef.Labels)) > 0:
		q = query{
			byLabels:    pointer.Deref(targetRef.Labels),
			sectionName: pointer.Deref(targetRef.SectionName),
		}
	default:
		q = query{
			byIdentifier: pointer.To(core_model.TargetRefToResourceIdentifier(tMeta, targetRef)),
			sectionName:  pointer.Deref(targetRef.SectionName),
		}
	}

	// backwards compatibility, we want old policies with targetRef{kind:MeshService,name:backend_kuma-demo_svc_8080}
	// to resolve to new MeshService backend in the kuma-demo namespace on port 8080
	if targetRef.Kind == common_api.MeshService && pointer.Deref(targetRef.SectionName) == "" {
		if name, namespace, port, err := parseService(pointer.Deref(targetRef.Name)); err == nil {
			q = query{
				byLabels: map[string]string{
					mesh_proto.KubeNamespaceTag: namespace,
					mesh_proto.DisplayName:      name,
				},
				port: port,
			}
		}
	}

	// resolve query without taking port/sectionName into account
	rtype := core_model.ResourceType(targetRef.Kind)
	var resources []core_model.Resource
	switch {
	case q.byLabels != nil:
		list := reader.ListOrEmpty(rtype).GetItems()
		trLabels := subsetutils.NewSubset(q.byLabels)
		for _, r := range list {
			rLabels := subsetutils.NewSubset(r.GetMeta().GetLabels())
			if trLabels.IsSubset(rLabels) {
				resources = append(resources, r)
			}
		}
	case q.byIdentifier != nil:
		if r := reader.Get(rtype, *q.byIdentifier); r != nil {
			resources = []core_model.Resource{r}
		}
	}

	if len(resources) == 0 {
		return []*ResourceSection{}
	}

	if q.port == 0 && q.sectionName == "" {
		result := make([]*ResourceSection, len(resources))
		for i := range resources {
			result[i] = &ResourceSection{Resource: resources[i]}
		}
		return result
	}

	// filter out resources that don't have requested section name or port
	var result []*ResourceSection
	for _, r := range resources {
		if resourceWithPorts, ok := r.(ResourceWithPorts); ok {
			if port := q.findPort(resourceWithPorts.GetPorts()); port != nil {
				result = append(result, &ResourceSection{
					Resource:    r,
					SectionName: port.GetNameOrStringifyPort(),
				})
			}
		} else {
			result = append(result, &ResourceSection{Resource: r})
		}
	}

	return result
}

func tryParsePort(s string) (uint32, bool) {
	u, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, false
	}
	return uint32(u), true
}

// parseService is copied from pkg/plugins/runtime/k8s/controllers/outbound_converter.go
// but when port is not specified it returns 0 instead of IANA Reserved port 49151.
// We don't need reserved port in the original 'parseService',
// there is an issue to fix it https://github.com/kumahq/kuma/issues/12834
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
		port = 0
	default:
		return "", "", 0, errors.Errorf("service tag in unexpected format")
	}

	name, namespace := segments[0], segments[1]
	return name, namespace, port, nil
}
