package v1alpha1

import (
	"fmt"
	"github.com/kumahq/kuma/pkg/util/validation"

	"github.com/asaskevich/govalidator"

	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshAccessLogResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.GetTargetRef()))
	if len(r.Spec.GetTo()) == 0 && len(r.Spec.GetFrom()) == 0 {
		verr.AddViolationAt(path, "at least one of 'from', 'to' has to be defined")
	}
	verr.AddErrorAt(path, validateTo(r.Spec.GetTo()))
	verr.AddErrorAt(path, validateFrom(r.Spec.GetFrom()))
	verr.AddErrorAt(path, validateIncompatibleCombinations(r.Spec))
	return verr.OrNil()
}
func validateTop(targetRef *common_proto.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			common_proto.TargetRef_Mesh,
			common_proto.TargetRef_MeshSubset,
			common_proto.TargetRef_MeshService,
			common_proto.TargetRef_MeshServiceSubset,
			common_proto.TargetRef_MeshGatewayRoute,
			common_proto.TargetRef_MeshHTTPRoute,
		},
	})
	return targetRefErr
}
func validateFrom(from []*MeshAccessLog_From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				common_proto.TargetRef_Mesh,
			},
		}))

		defaultField := path.Field("default")
		if fromItem.GetDefault() == nil {
			verr.AddViolationAt(defaultField, "must be defined")
		} else {
			verr.AddErrorAt(defaultField, validateDefault(fromItem.Default))
		}
	}
	return verr
}
func validateTo(to []*MeshAccessLog_To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(toItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_proto.TargetRef_Kind{
				common_proto.TargetRef_Mesh,
				common_proto.TargetRef_MeshService,
			},
		}))

		defaultField := path.Field("default")
		if toItem.GetDefault() == nil {
			verr.AddViolationAt(defaultField, "must be defined")
		} else {
			verr.AddErrorAt(defaultField, validateDefault(toItem.Default))
		}
	}
	return verr
}

func validateDefault(conf *MeshAccessLog_Conf) validators.ValidationError {
	var verr validators.ValidationError
	for backendIdx, backend := range conf.Backends {
		verr.AddErrorAt(validators.RootedAt("backends").Index(backendIdx), validateBackend(backend))
	}
	return verr
}

func validateBackend(backend *MeshAccessLog_Backend) validators.ValidationError {
	var verr validators.ValidationError
	file := validation.Bool2Int(backend.GetFile() != nil)
	tcp := validation.Bool2Int(backend.GetTcp() != nil)

	if file+tcp != 1 {
		verr.AddViolation("", validation.MustHaveOnlyOneMessage("backend", "tcp", "file"))
	}

	verr.AddErrorAt(validators.RootedAt("file").Field("format"), validateFormat(backend.GetFile().GetFormat()))
	verr.AddErrorAt(validators.RootedAt("tcp").Field("format"), validateFormat(backend.GetTcp().GetFormat()))

	if backend.GetFile() != nil {
		isFilePath, _ := govalidator.IsFilePath(backend.GetFile().GetPath())
		if !isFilePath {
			verr.AddViolationAt(validators.RootedAt("file").Field("path"), `file backend requires a valid path`)
		}
	}

	if backend.GetTcp() != nil {
		if !govalidator.IsURL(backend.GetTcp().GetAddress()) {
			verr.AddViolationAt(validators.RootedAt("tcp").Field("address"), `tcp backend requires valid address`)
		}
	}
	return verr
}

func validateFormat(format *MeshAccessLog_Format) validators.ValidationError {
	var verr validators.ValidationError
	if format == nil {
		return verr
	}
	plain := validation.Bool2Int(format.GetPlain() != "")
	json := validation.Bool2Int(format.GetJson() != nil)

	if plain+json != 1 {
		verr.AddViolation("", validation.MustHaveOnlyOneMessage("format", "plain", "json"))
	}

	if format.GetJson() != nil {
		for idx, field := range format.GetJson() {
			indexedField := validators.RootedAt("json").Index(idx)
			if field.GetKey() == "" {
				verr.AddViolationAt(indexedField.Field("key"), `key cannot be empty`)
			}
			if field.GetValue() == "" {
				verr.AddViolationAt(indexedField.Field("value"), `value cannot be empty`)
			}
			if !govalidator.IsJSON(fmt.Sprintf(`{"%s": "%s"}`, field.GetKey(), field.GetValue())) {
				verr.AddViolationAt(indexedField, `is not a valid JSON object`)
			}
		}
	}
	return verr
}

func validateIncompatibleCombinations(spec *MeshAccessLog) validators.ValidationError {
	var verr validators.ValidationError
	targetRef := spec.GetTargetRef().GetKindEnum()
	if targetRef == common_proto.TargetRef_MeshGatewayRoute && len(spec.GetTo()) > 0 {
		verr.AddViolation("to", `cannot use "to" when "targetRef" is "MeshGatewayRoute" - there is no outbound`)
	}
	if targetRef == common_proto.TargetRef_MeshHTTPRoute && len(spec.GetTo()) > 0 {
		verr.AddViolation("to", `cannot use "to" when "targetRef" is "MeshHTTPRoute" - "to" always goes to the application`)
	}
	return verr
}
