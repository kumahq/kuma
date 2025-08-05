package resolve

import (
	"fmt"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/util/pointer"
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

	rr := &RealResourceBackendRef{
		Resource: TargetRefToKRI(origin, br.TargetRef),
		Origin:   origin,
		Weight:   pointer.DerefOr(br.Weight, 1),
	}

	if labels := pointer.Deref(br.Labels); len(labels) > 0 {
		rr.Resource = resolver(core_model.ResourceType(br.Kind), labels)
	}

	if rr.Resource.IsEmpty() {
		return ResolvedBackendRef{}, false
	}

	if br.Port != nil {
		rr.Resource.SectionName = fmt.Sprintf("%d", *br.Port)
	}

	return ResolvedBackendRef{Ref: rr}, true
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

func TargetRefToKRI(origin kri.Identifier, ref common_api.TargetRef) kri.Identifier {
	if origin.IsEmpty() {
		return kri.Identifier{}
	}

	if ref.Kind == common_api.Mesh {
		return kri.Identifier{
			ResourceType: core_mesh.MeshType,
			Name:         origin.Mesh,
		}
	}

	return kri.Identifier{
		ResourceType: core_model.ResourceType(ref.Kind),
		Mesh:         origin.Mesh,
		Zone:         origin.Zone,
		Namespace:    pointer.DerefOr(ref.Namespace, origin.Namespace),
		Name:         pointer.Deref(ref.Name),
	}
}
