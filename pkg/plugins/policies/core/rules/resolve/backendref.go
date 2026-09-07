package resolve

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
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
	if !br.ReferencesRealObject() {
		return ResolvedBackendRef{}, false
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
		return ResolvedBackendRef{}, false
	}
	rr.Resource.SectionName = sectionName

	return ResolvedBackendRef{Ref: rr}, true
}

type IsResolvedBackendRef interface {
	isResolvedBackendRef()
}

type ResolvedBackendRef struct {
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

func (rbr *ResolvedBackendRef) RealResourceBackendRef() *RealResourceBackendRef {
	if rr, ok := rbr.Ref.(*RealResourceBackendRef); ok {
		return rr
	}
	return nil
}

type RealResourceBackendRef struct {
	Resource kri.Identifier
	Origin   kri.Identifier
	Weight   uint
}

func (rbr *RealResourceBackendRef) isResolvedBackendRef() {}
