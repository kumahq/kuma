package resolve

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type LabelResourceIdentifierResolver func(core_model.ResourceType, map[string]string) *kri.Identifier

func BackendRefOrNil(origin *kri.Identifier, br common_api.BackendRef, resolver LabelResourceIdentifierResolver) *ResolvedBackendRef {
	if br, ok := BackendRef(origin, br, resolver); ok {
		return &br
	}
	return nil
}

func BackendRef(origin *kri.Identifier, br common_api.BackendRef, resolver LabelResourceIdentifierResolver) (ResolvedBackendRef, bool) {
	switch {
	case br.Kind == common_api.MeshService && br.ReferencesRealObject():
	case br.Kind == common_api.MeshExternalService:
	case br.Kind == common_api.MeshMultiZoneService:
	default:
		return ResolvedBackendRef{Ref: pointer.To(LegacyBackendRef(br))}, true
	}

	rr := RealResourceBackendRef{
		Origin: origin,
		Weight: pointer.DerefOr(br.Weight, 1),
	}

	if labels := pointer.Deref(br.Labels); len(labels) > 0 {
		rr.Resource = resolver(core_model.ResourceType(br.Kind), labels)
	} else if origin != nil {
		rr.Resource = pointer.To(targetRefToKRI(br.TargetRef, origin.Mesh, origin.Zone, origin.Namespace))
	}

	if rr.Resource == nil {
		return ResolvedBackendRef{}, false
	}

	if br.Port != nil {
		rr.Resource.SectionName = fmt.Sprintf("%d", *br.Port)
	}

	return ResolvedBackendRef{Ref: &rr}, true
}

type IsResolvedBackendRef interface {
	isResolvedBackendRef()
}

type ResolvedBackendRef struct {
	// Ref is either LegacyBackendRef or RealResourceBackendRef
	Ref IsResolvedBackendRef
}

func NewResolvedBackendRef(r IsResolvedBackendRef) *ResolvedBackendRef {
	return &ResolvedBackendRef{Ref: r}
}

func (rbr *ResolvedBackendRef) ReferencesRealResource() bool {
	if rbr == nil {
		return false
	}
	if rbr.Ref == nil {
		return false
	}
	_, ok := rbr.Ref.(*RealResourceBackendRef)
	return ok
}

func (rbr *ResolvedBackendRef) ResourceOrNil() *kri.Identifier {
	if rr := rbr.RealResourceBackendRef(); rr != nil {
		return rr.Resource
	}
	return nil
}

func (rbr *ResolvedBackendRef) LegacyBackendRef() *LegacyBackendRef {
	if lbr, ok := rbr.Ref.(*LegacyBackendRef); ok {
		return lbr
	}
	return nil
}

func (rbr *ResolvedBackendRef) RealResourceBackendRef() *RealResourceBackendRef {
	if rr, ok := rbr.Ref.(*RealResourceBackendRef); ok {
		return rr
	}
	return nil
}

type LegacyBackendRef common_api.BackendRef

func (lbr *LegacyBackendRef) isResolvedBackendRef() {}

type RealResourceBackendRef struct {
	Resource *kri.Identifier
	Origin   *kri.Identifier
	Weight   uint
}

func (rbr *RealResourceBackendRef) isResolvedBackendRef() {}

func TargetRefToKRI(meta core_model.ResourceMeta, tr common_api.TargetRef) kri.Identifier {
	return targetRefToKRI(
		tr,
		meta.GetMesh(),
		meta.GetLabels()[mesh_proto.ZoneTag],
		meta.GetLabels()[mesh_proto.KubeNamespaceTag],
	)
}

func targetRefToKRI(tr common_api.TargetRef, mesh, zone, fallbackNamespace string) kri.Identifier {
	if tr.Kind == common_api.Mesh {
		return kri.Identifier{
			ResourceType: core_model.ResourceType(tr.Kind),
			Name:         mesh,
		}
	}

	return kri.Identifier{
		ResourceType: core_model.ResourceType(tr.Kind),
		Mesh:         mesh,
		Zone:         zone,
		Namespace:    pointer.DerefOr(tr.Namespace, fallbackNamespace),
		Name:         pointer.Deref(tr.Name),
	}
}
