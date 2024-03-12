package v1alpha1

import (
	"github.com/asaskevich/govalidator"
	"regexp"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshMetricResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := mesh.ValidateTargetRef(targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh, common_api.MeshSubset, common_api.MeshService, common_api.MeshServiceSubset, common_api.MeshGateway,
		},
		GatewayListenerTagsAllowed: true,
	})
	return targetRefErr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	verr.AddError("sidecar", validateSidecar(conf.Sidecar))
	verr.AddError("applications", validateApplications(conf.Applications))
	verr.AddError("backends", validateBackend(conf.Backends))
	return verr
}

func validateSidecar(sidecar *Sidecar) validators.ValidationError {
	var verr validators.ValidationError
	if sidecar == nil {
		return verr
	}

	if sidecar.Profiles != nil {
		profiles := sidecar.Profiles
		path := validators.RootedAt("profiles")

		if profiles.Exclude != nil {
			verr.AddError(path.Field("exclude").String(), validateSelector(*profiles.Exclude))
		}
		if profiles.Include != nil {
			verr.AddError(path.Field("include").String(), validateSelector(*profiles.Include))
		}
	}

	return verr
}

func validateSelector(selector Selector) validators.ValidationError {
	var verr validators.ValidationError

	switch selector.Type {
	case RegexSelectorType:
		_, err := regexp.Compile(selector.Match)
		if err != nil {
			verr.AddViolation("match", "invalid regex")
		}
	case PrefixSelectorType,ExactSelectorType:
	default:
		verr.AddViolation("match", "unrecognized type")
	}

	return verr
}

func validateApplications(applications *[]Application) validators.ValidationError {
	var verr validators.ValidationError
	if applications == nil {
		return verr
	}

	for idx, application := range *applications {
		verr.Add(validators.ValidatePort(validators.RootedAt("application").Index(idx), application.Port))
	}

	return verr
}

func validateBackend(backends *[]Backend) validators.ValidationError {
	var verr validators.ValidationError
	if backends == nil {
		return verr
	}

	for idx, backend := range *backends {
		path := validators.RootedAt("backend").Index(idx)
		switch backend.Type {
		case PrometheusBackendType:
			if backend.Prometheus == nil {
				verr.AddViolationAt(path.Field("prometheus"), validators.MustBeDefined)
			} else {
				verr.Add(validators.ValidatePort(path.Field("port"), backend.Prometheus.Port))
			}
		case OpenTelemetryBackendType:
			if backend.OpenTelemetry == nil {
				verr.AddViolationAt(path.Field("openTelemetry"), validators.MustBeDefined)
			} else if !govalidator.IsURL(backend.OpenTelemetry.Endpoint) {
				verr.AddViolationAt(path.Field("openTelemetry").Field("endpoint"), "must be a valid url")
			}
		default:
			verr.AddViolationAt(path, "unrecognized type")
		}
	}

	return verr
}
