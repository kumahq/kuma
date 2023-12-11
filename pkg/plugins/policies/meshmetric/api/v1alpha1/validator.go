package v1alpha1

import (
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
			common_api.Mesh, common_api.MeshSubset, common_api.MeshService, common_api.MeshServiceSubset,
		},
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

	if sidecar.Regex != nil {
		_, err := regexp.Compile(*sidecar.Regex)
		if err != nil {
			verr.AddViolation("regex", "invalid regex")
		}
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
		default:
			verr.AddViolationAt(path, "unrecognized type")
		}
	}

	return verr
}
