package v1alpha1

import (
	"fmt"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshAccessLogResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.GetTargetRef()))
	if len(r.Spec.To) == 0 && len(r.Spec.From) == 0 {
		verr.AddViolationAt(path, "at least one of 'from', 'to' has to be defined")
	}
	verr.AddErrorAt(path, validateTo(r.Spec.To))
	verr.AddErrorAt(path, validateFrom(r.Spec.From))
	verr.AddErrorAt(path, validateIncompatibleCombinations(r.Spec))
	return verr.OrNil()
}
func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
			common_api.MeshGatewayRoute,
		},
	})
	return targetRefErr
}
func validateFrom(from []From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		}))

		defaultField := path.Field("default")
		verr.AddErrorAt(defaultField, validateDefault(fromItem.Default))
	}
	return verr
}
func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(toItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
			},
		}))

		defaultField := path.Field("default")
		verr.AddErrorAt(defaultField, validateDefault(toItem.Default))
	}
	return verr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	if conf.Backends == nil {
		verr.AddViolation("backends", validators.MustBeDefined)
		return verr
	}
	for backendIdx, backend := range *conf.Backends {
		verr.AddErrorAt(validators.RootedAt("backends").Index(backendIdx), validateBackend(backend))
	}
	return verr
}

func validateBackend(backend Backend) validators.ValidationError {
	var verr validators.ValidationError

	if (backend.File != nil) == (backend.Tcp != nil) {
		verr.AddViolation("", validators.MustHaveOnlyOne("backend", "tcp", "file"))
	}

	switch {
	case backend.File != nil:
		if backend.File.Format != nil {
			verr.AddErrorAt(validators.RootedAt("file").Field("format"), validateFormat(*backend.File.Format))
		}
		isFilePath, _ := govalidator.IsFilePath(backend.File.Path)
		if !isFilePath {
			verr.AddViolationAt(validators.RootedAt("file").Field("path"), `file backend requires a valid path`)
		}
	case backend.Tcp != nil:
		if backend.Tcp.Format != nil {
			verr.AddErrorAt(validators.RootedAt("tcp").Field("format"), validateFormat(*backend.Tcp.Format))
		}
		if !govalidator.IsURL(backend.Tcp.Address) {
			verr.AddViolationAt(validators.RootedAt("tcp").Field("address"), `tcp backend requires valid address`)
		}
	}

	return verr
}

func validateFormat(format Format) validators.ValidationError {
	var verr validators.ValidationError

	if (format.Plain != nil) == (format.Json != nil) {
		verr.AddViolation("", validators.MustHaveOnlyOne("format", "plain", "json"))
	}

	switch {
	case format.Plain != nil:
		if *format.Plain == "" {
			verr.AddViolation("plain", validators.MustNotBeEmpty)
		}
	case format.Json != nil:
		if len(*format.Json) == 0 {
			verr.AddViolation("json", validators.MustNotBeEmpty)
		}
		for idx, field := range *format.Json {
			path := validators.RootedAt("json").Index(idx)
			if field.Key == "" {
				verr.AddViolationAt(path.Field("key"), `key cannot be empty`)
			}
			if field.Value == "" {
				verr.AddViolationAt(path.Field("value"), `value cannot be empty`)
			}
			if !govalidator.IsJSON(fmt.Sprintf(`{"%s": "%s"}`, field.Key, field.Value)) {
				verr.AddViolationAt(path, `is not a valid JSON object`)
			}
		}
	}

	return verr
}

func validateIncompatibleCombinations(spec *MeshAccessLog) validators.ValidationError {
	var verr validators.ValidationError
	targetRef := spec.GetTargetRef().Kind
	if targetRef == common_api.MeshGatewayRoute && len(spec.To) > 0 {
		verr.AddViolation("to", `cannot use "to" when "targetRef" is "MeshGatewayRoute" - there is no outbound`)
	}
	return verr
}
