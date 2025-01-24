package v1alpha1

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/asaskevich/govalidator"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshMetricResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}

func (r *MeshMetricResource) validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	switch core_model.PolicyRole(r.GetMeta()) {
	case mesh_proto.SystemPolicyRole:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
				common_api.MeshGateway,
				common_api.Dataplane,
			},
			GatewayListenerTagsAllowed: true,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.Dataplane,
				common_api.MeshServiceSubset,
			},
		})
	}
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
		if profiles.Exclude != nil {
			verr.AddError("profiles", validateSelectors(*profiles.Exclude, "exclude"))
		}
		if profiles.Include != nil {
			verr.AddError("profiles", validateSelectors(*profiles.Include, "include"))
		}
		if profiles.AppendProfiles != nil {
			verr.AddError("profiles.appendProfiles", validateAppendProfiles(*profiles.AppendProfiles))
		}
	}
	return verr
}

func validateAppendProfiles(profiles []Profile) validators.ValidationError {
	var verr validators.ValidationError
	for i, profile := range profiles {
		path := validators.Root().Index(i)

		switch profile.Name {
		case AllProfileName, NoneProfileName, BasicProfileName:
		default:
			verr.AddViolation(path.Field("name").String(), fmt.Sprintf("unrecognized profile name '%s' - 'All', 'None', 'Basic' are supported", profile.Name))
		}
	}
	return verr
}

func validateSelectors(selectors []Selector, selectorType string) validators.ValidationError {
	var verr validators.ValidationError

	for i, selector := range selectors {
		path := validators.RootedAt(selectorType).Index(i)
		switch selector.Type {
		case RegexSelectorType:
			_, err := regexp.Compile(selector.Match)
			if err != nil {
				verr.AddViolation(path.Field("match").String(), "invalid regex")
			}
		case PrefixSelectorType, ExactSelectorType, ContainsSelectorType:
		default:
			verr.AddViolation(path.Field("type").String(), fmt.Sprintf("unrecognized type '%s' - 'Regex', 'Prefix', 'Exact' are supported", selector.Type))
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
		case OpenTelemetryBackendType:
			if backend.OpenTelemetry == nil {
				verr.AddViolationAt(path.Field("openTelemetry"), validators.MustBeDefined)
			} else {
				endpoint := backend.OpenTelemetry.Endpoint
				if !govalidator.IsURL(endpoint) {
					verr.AddViolationAt(path.Field("openTelemetry").Field("endpoint"), "must be a valid url")
				}
				if strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
					verr.AddViolationAt(path.Field("openTelemetry").Field("endpoint"), "must not use schema")
				}
			}
		default:
			verr.AddViolationAt(path, "unrecognized type")
		}
	}

	return verr
}
