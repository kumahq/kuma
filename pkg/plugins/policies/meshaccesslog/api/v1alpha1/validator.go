package v1alpha1

import (
	"fmt"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
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
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshGateway,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
		GatewayListenerTagsAllowed: true,
	})
	return targetRefErr
}

func validateFrom(from []From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(fromItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
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
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(toItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
				common_api.MeshExternalService,
				common_api.MeshMultiZoneService,
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

	switch backend.Type {
	case FileBackendType:
		root := validators.RootedAt("file")
		if backend.File == nil {
			verr.AddViolationAt(root, validators.MustBeDefined)
			break
		}

		if backend.File.Format != nil {
			verr.AddErrorAt(root.Field("format"), validateFormat(*backend.File.Format))
		}
		isFilePath, _ := govalidator.IsFilePath(backend.File.Path)
		if !isFilePath {
			verr.AddViolationAt(root.Field("path"), `file backend requires a valid path`)
		}
	case TCPBackendType:
		root := validators.RootedAt("tcp")
		if backend.Tcp == nil {
			verr.AddViolationAt(root, validators.MustBeDefined)
			break
		}

		if backend.Tcp.Format != nil {
			verr.AddErrorAt(root.Field("format"), validateFormat(*backend.Tcp.Format))
		}
		if !govalidator.IsURL(backend.Tcp.Address) {
			verr.AddViolationAt(root.Field("address"), `tcp backend requires valid address`)
		}
	case OtelTelemetryBackendType:
		root := validators.RootedAt("openTelemetry")
		if backend.OpenTelemetry == nil {
			verr.AddViolationAt(root, validators.MustBeDefined)
			break
		}
	default:
		panic(fmt.Sprintf("unknown backend type %v", backend.Type))
	}

	return verr
}

func validateFormat(format Format) validators.ValidationError {
	var verr validators.ValidationError

	switch format.Type {
	case PlainFormatType:
		root := validators.RootedAt("plain")
		if format.Plain == nil {
			verr.AddViolationAt(root, validators.MustBeDefined)
			break
		}
		if *format.Plain == "" {
			verr.AddViolationAt(root, validators.MustNotBeEmpty)
		}
	case JsonFormatType:
		root := validators.RootedAt("json")
		if format.Json == nil {
			verr.AddViolationAt(root, validators.MustBeDefined)
			break
		}

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
	default:
		panic(fmt.Sprintf("unknown backend type %v", format.Type))
	}

	return verr
}
