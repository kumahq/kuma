package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshTimeoutResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef))
	if len(r.Spec.To) == 0 && len(r.Spec.From) == 0 {
		verr.AddViolationAt(path, "at least one of 'from', 'to' has to be defined")
	}
	verr.AddErrorAt(path, validateFrom(r.Spec.From, r.Spec.TargetRef.Kind))
	verr.AddErrorAt(path, validateTo(r.Spec.To, r.Spec.TargetRef.Kind))
	return verr.OrNil()
}

func (r *MeshTimeoutResource) validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	switch core_model.PolicyRole(r.GetMeta()) {
	case mesh_proto.SystemPolicyRole:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshGateway,
				common_api.MeshService,
				common_api.MeshServiceSubset,
				common_api.MeshHTTPRoute,
			},
			GatewayListenerTagsAllowed: true,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
		})
	}
}

func validateFrom(from []From, topLevelKind common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError
	fromPath := validators.RootedAt("from")

	switch topLevelKind {
	case common_api.MeshHTTPRoute, common_api.MeshGateway:
		if len(from) != 0 {
			verr.AddViolationAt(fromPath, validators.MustNotBeDefined)
		}
		return verr
	}

	for idx, fromItem := range from {
		path := fromPath.Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(fromItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		}))

		defaultField := path.Field("default")
		verr.Add(validateDefault(defaultField, fromItem.Default, topLevelKind))
	}
	return verr
}

func validateTo(to []To, topLevelKind common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)

		var supportedKinds []common_api.TargetRefKind
		switch topLevelKind {
		case common_api.MeshHTTPRoute, common_api.MeshGateway:
			supportedKinds = []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshExternalService,
			}
		default:
			supportedKinds = []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
				common_api.MeshExternalService,
				common_api.MeshMultiZoneService,
			}
		}

		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(toItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: supportedKinds,
		}))

		defaultField := path.Field("default")
		verr.Add(validateDefault(defaultField, toItem.Default, topLevelKind))
	}
	return verr
}

func validateDefault(path validators.PathBuilder, conf Conf, topLevelKind common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError

	if conf.ConnectionTimeout == nil && conf.IdleTimeout == nil && conf.Http == nil {
		verr.AddViolationAt(path, "at least one timeout should be configured")
		return verr
	}

	if topLevelKind == common_api.MeshHTTPRoute {
		msg := "can't be specified when top-level TargetRef is referencing MeshHTTPRoute"

		verr.Add(validators.ValidateNil(path.Field("connectionTimeout"), conf.ConnectionTimeout, msg))
		verr.Add(validators.ValidateNil(path.Field("idleTimeout"), conf.IdleTimeout, msg))
		if http := conf.Http; http != nil {
			httpPath := path.Field("http")
			verr.Add(validators.ValidateNil(httpPath.Field("maxStreamDuration"), http.MaxStreamDuration, msg))
			verr.Add(validators.ValidateNil(httpPath.Field("maxConnectionDuration"), http.MaxConnectionDuration, msg))
		}
	}

	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("connectionTimeout"), conf.ConnectionTimeout))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("idleTimeout"), conf.IdleTimeout))

	verr.Add(validateHttp(path.Field("http"), conf.Http))
	return verr
}

func validateHttp(path validators.PathBuilder, http *Http) validators.ValidationError {
	var verr validators.ValidationError
	if http == nil {
		return verr
	}

	if http.RequestTimeout == nil && http.StreamIdleTimeout == nil && http.MaxStreamDuration == nil && http.MaxConnectionDuration == nil {
		verr.AddViolationAt(path, "at least one timeout in this section should be configured")
	}

	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("requestTimeout"), http.RequestTimeout))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("streamIdleTimeout"), http.StreamIdleTimeout))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("maxStreamDuration"), http.MaxStreamDuration))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("maxConnectionDuration"), http.MaxConnectionDuration))
	verr.Add(validators.ValidateDurationNotNegativeOrNil(path.Field("requestHeadersTimeout"), http.RequestHeadersTimeout))

	return verr
}
