package v1alpha1

import (
	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshAccessLogResource) validate() error {
	var verr validators.ValidationError
	spec := validators.RootedAt("spec")

	r.validateTop(spec, &verr)
	r.validateFrom(spec, &verr)
	r.validateTo(spec, &verr)
	r.validateIncompatibleCombinations(spec, &verr)
	r.validateToOrFromDefined(spec, &verr)


	return verr.OrNil()
}

func (r *MeshAccessLogResource) validateBackend(backend *MeshAccessLog_Backend, verr *validators.ValidationError, backendIndexed validators.PathBuilder) {
	reference := bool2int(backend.GetReference() != nil)
	file := bool2int(backend.GetFile() != nil)
	tcp := bool2int(backend.GetTcp() != nil)

	if reference+file+tcp != 1 {
		verr.AddViolationAt(backendIndexed, `backend can have only one type type defined: tcp, file, reference`)
	}

	r.validateFormats(backend, verr, backendIndexed)

	if backend.GetFile() != nil && backend.GetFile().Path == "" {
		verr.AddViolationAt(backendIndexed.Field("file").Field("path"), `file backend requires a path`)
	}
}

func (r *MeshAccessLogResource) validateFormats(backend *MeshAccessLog_Backend, verr *validators.ValidationError, backendIndexed validators.PathBuilder) {
	var formats []*MeshAccessLog_Format
	if backend.GetFile() != nil {
		formats = append(formats, backend.GetFile().Format)
	}
	if backend.GetTcp() != nil {
		formats = append(formats, backend.GetTcp().Format)
	}
	for _, format := range formats {
		plain := bool2int(format.GetPlain() != "")
		json := bool2int(format.GetJson() != nil)

		if plain+json > 1 {
			verr.AddViolationAt(backendIndexed, `format can only have one type defined: plain, json`)
		}

		if format.GetJson() != nil {
			for idx, field := range format.GetJson() {
				indexedField := backendIndexed.Field("json").Index(idx)
				if field.GetKey() == "" {
					verr.AddViolationAt(indexedField.Field("key"), `key cannot be empty`)
				}
				if field.GetValue() == "" {
					verr.AddViolationAt(indexedField.Field("value"), `value cannot be empty`)
				}
			}
		}
	}
}

func (r *MeshAccessLogResource) validateToOrFromDefined(spec validators.PathBuilder, verr *validators.ValidationError) {
	if len(r.Spec.GetFrom()) == 0 && len(r.Spec.GetTo()) == 0 {
		verr.AddViolationAt(spec, `at lest one of "from", "to" has to be defined`)
	}
}

func (r *MeshAccessLogResource) validateIncompatibleCombinations(spec validators.PathBuilder, verr *validators.ValidationError) {
	to := spec.Field("to")
	targetRef := r.Spec.GetTargetRef().GetKindEnum()
	if targetRef == common_proto.TargetRef_MeshGatewayRoute && len(r.Spec.GetTo()) > 0 {
		verr.AddViolationAt(to.Index(0), `cannot use "to" when "targetRef" is "MeshGatewayRoute" - there is no outbound`)
	}
	if targetRef == common_proto.TargetRef_MeshHTTPRoute && len(r.Spec.GetTo()) > 0 {
		verr.AddViolationAt(to.Index(0), `cannot use "to" when "targetRef" is "MeshHTTPRoute" - "to" always goes to the application`)
	}
}

func (r *MeshAccessLogResource) validateTo(spec validators.PathBuilder, verr *validators.ValidationError) {
	to := spec.Field("to")
	for idx, toItem := range r.Spec.GetTo() {
		targetRefErr := matcher_validators.ValidateTargetRef(to.Index(idx).Field("targetRef"), toItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				common_proto.TargetRef_Mesh,
				common_proto.TargetRef_MeshService,
			},
		})
		verr.AddError("", targetRefErr)

		if toItem.GetDefault() == nil {
			verr.AddViolationAt(to.Index(idx).Field("default"), "cannot be nil")
		}

		toIndexed := to.Index(idx)
		for backendIdx, backend := range toItem.GetDefault().GetBackends() {
			backendIndexed := toIndexed.Field("default").Field("backend").Index(backendIdx)
			r.validateBackend(backend, verr, backendIndexed)
		}
	}
}

func (r *MeshAccessLogResource) validateFrom(spec validators.PathBuilder, verr *validators.ValidationError) {
	from := spec.Field("from")
	for idx, fromItem := range r.Spec.GetFrom() {
		targetRefErr := matcher_validators.ValidateTargetRef(from.Index(idx).Field("targetRef"), fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				common_proto.TargetRef_Mesh,
			},
		})
		verr.AddError("", targetRefErr)

		if fromItem.GetDefault() == nil {
			verr.AddViolationAt(from.Index(idx).Field("default"), "cannot be nil")
		}

		toIndexed := from.Index(idx)
		for backendIdx, backend := range fromItem.GetDefault().GetBackends() {
			backendIndexed := toIndexed.Field("default").Field("backend").Index(backendIdx)
			r.validateBackend(backend, verr, backendIndexed)
		}
	}
}

func (r *MeshAccessLogResource) validateTop(spec validators.PathBuilder, verr *validators.ValidationError) {
	top := spec.Field("targetRef")
	targetRefErr := matcher_validators.ValidateTargetRef(top, r.Spec.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			common_proto.TargetRef_Mesh,
			common_proto.TargetRef_MeshSubset,
			common_proto.TargetRef_MeshService,
			common_proto.TargetRef_MeshServiceSubset,
			common_proto.TargetRef_MeshGatewayRoute,
			common_proto.TargetRef_MeshHTTPRoute,
		},
	})
	verr.AddError("", targetRefErr)
}

func bool2int(b bool) int {
	if b {
		return 1
	}
	return 0
}
