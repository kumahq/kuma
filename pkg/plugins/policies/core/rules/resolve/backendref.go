package resolve

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

type LabelResourceIdentifierResolver func(core_model.ResourceType, map[string]string) kri.Identifier

func BackendRefOrNil(origin kri.Identifier, br common_api.BackendRef, resolver LabelResourceIdentifierResolver) *ResolvedBackendRef {
	if br, ok := BackendRef(origin, br, resolver); ok {
		return &br
	}
	return nil
}

func BackendRef(origin kri.Identifier, br common_api.BackendRef, resolver LabelResourceIdentifierResolver) (ResolvedBackendRef, bool) {
	switch {
	case br.Kind == common_api.MeshService && br.ReferencesRealObject():
	case br.Kind == common_api.MeshExternalService:
	case br.Kind == common_api.MeshMultiZoneService:
	default:
		return ResolvedBackendRef{Ref: pointer.To(LegacyBackendRef(br))}, true
	}

	labels, sectionName, ok := br.RealResourceSelector(origin.Namespace)
	if !ok {
		return ResolvedBackendRef{}, false
	}

	rr := &RealResourceBackendRef{
		Resource: resolver(core_model.ResourceType(br.Kind), labels),
		Origin:   origin,
		Weight:   pointer.DerefOr(br.Weight, 1),
	}
	if rr.Resource.IsEmpty() {
		if shouldFallbackToLegacyMeshService(br) {
			return ResolvedBackendRef{Ref: pointer.To(LegacyBackendRef(br))}, true
		}
		return ResolvedBackendRef{}, false
	}
	rr.Resource.SectionName = sectionName

	return ResolvedBackendRef{Ref: rr}, true
}

func shouldFallbackToLegacyMeshService(br common_api.BackendRef) bool {
	labels := pointer.Deref(br.Labels)
	return br.Kind == common_api.MeshService &&
		labels[mesh_proto.DisplayName] != "" &&
		len(labels) == 1 &&
		br.SectionName == nil &&
		br.Port == nil
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

func (rbr *ResolvedBackendRef) Resource() kri.Identifier {
	if rr := rbr.RealResourceBackendRef(); rr != nil {
		return rr.Resource
	}
	return kri.Identifier{}
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
	Resource kri.Identifier
	Origin   kri.Identifier
	Weight   uint
}

func (rbr *RealResourceBackendRef) isResolvedBackendRef() {}
