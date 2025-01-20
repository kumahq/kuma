package common

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
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

func ResolveTargetRef(targetRef common_api.TargetRef, tMeta core_model.ResourceMeta, reader ResourceReader) []*ResourceSection {
	if !targetRef.Kind.IsRealResource() {
		return nil
	}
	rtype := core_model.ResourceType(targetRef.Kind)
	list := reader.ListOrEmpty(rtype).GetItems()

	var implicitPort uint32
	implicitLabels := map[string]string{}
	if targetRef.Kind == common_api.MeshService && targetRef.SectionName == "" {
		if name, namespace, port, err := parseService(targetRef.Name); err == nil {
			implicitLabels[mesh_proto.KubeNamespaceTag] = namespace
			implicitLabels[mesh_proto.DisplayName] = name
			implicitPort = port
		}
	}

	labels := targetRef.Labels
	if len(implicitLabels) > 0 {
		labels = implicitLabels
	}

	if len(labels) > 0 {
		var rv []*ResourceSection
		trLabels := subsetutils.NewSubset(labels)
		for _, r := range list {
			rLabels := subsetutils.NewSubset(r.GetMeta().GetLabels())
			var implicitSectionName string
			if ms, ok := r.(*meshservice_api.MeshServiceResource); ok && implicitPort != 0 {
				for _, port := range ms.Spec.Ports {
					if port.Port == implicitPort {
						implicitSectionName = port.Name
					}
				}
			}
			sn := targetRef.SectionName
			if sn == "" {
				sn = implicitSectionName
			}
			if trLabels.IsSubset(rLabels) {
				rv = append(rv, &ResourceSection{
					Resource:    r,
					SectionName: sn,
				})
			}
		}
		return rv
	}

	ri := core_model.TargetRefToResourceIdentifier(tMeta, targetRef)
	if resource := reader.Get(rtype, ri); resource != nil {
		return []*ResourceSection{{
			Resource:    resource,
			SectionName: targetRef.SectionName,
		}}
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
