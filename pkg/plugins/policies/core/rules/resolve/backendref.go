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

func BackendRefOrNil(meta core_model.ResourceMeta, br common_api.BackendRef, resolver LabelResourceIdentifierResolver) *ResolvedBackendRef {
	if br, ok := BackendRef(meta, br, resolver); ok {
		return &br
	}
	return nil
}

func BackendRef(meta core_model.ResourceMeta, br common_api.BackendRef, resolver LabelResourceIdentifierResolver) (ResolvedBackendRef, bool) {
	switch {
	case br.Kind == common_api.MeshService && br.ReferencesRealObject():
	case br.Kind == common_api.MeshExternalService:
	case br.Kind == common_api.MeshMultiZoneService:
	default:
		return ResolvedBackendRef{Ref: pointer.To(LegacyBackendRef(br))}, true
	}

	rr := RealResourceBackendRef{
		Resource: pointer.To(TargetRefToKRI(meta, br.TargetRef)),
		Weight:   pointer.DerefOr(br.Weight, 1),
	}

	if len(pointer.Deref(br.Labels)) > 0 {
		ri := resolver(core_model.ResourceType(br.Kind), pointer.Deref(br.Labels))
		if ri == nil {
			return ResolvedBackendRef{}, false
		}
		rr.Resource = ri
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
	Weight   uint
}

func (rbr *RealResourceBackendRef) isResolvedBackendRef() {}

func TargetRefToKRI(meta core_model.ResourceMeta, tr common_api.TargetRef) kri.Identifier {
	switch tr.Kind {
	case common_api.Mesh:
		return kri.Identifier{
			ResourceType: core_model.ResourceType(tr.Kind),
			Name:         meta.GetMesh(),
		}
	default:
		var namespace string
		if pointer.Deref(tr.Namespace) != "" {
			namespace = pointer.Deref(tr.Namespace)
		} else {
			namespace = meta.GetLabels()[mesh_proto.KubeNamespaceTag]
		}
		return kri.Identifier{
			ResourceType: core_model.ResourceType(tr.Kind),
			Mesh:         meta.GetMesh(),
			Zone:         meta.GetLabels()[mesh_proto.ZoneTag],
			Namespace:    namespace,
			Name:         pointer.Deref(tr.Name),
		}
	}
}
