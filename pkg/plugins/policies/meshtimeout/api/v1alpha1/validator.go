package v1alpha1

import (
	"fmt"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (r *MeshTimeoutResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef, inbound.AffectsInbounds(r.Spec)))
	if len(pointer.Deref(r.Spec.Rules)) > 0 && len(pointer.Deref(r.Spec.To)) > 0 {
		verr.AddViolationAt(path, "fields 'to' must be empty when 'rules' is defined")
	}
	if len(pointer.Deref(r.Spec.Rules)) == 0 && len(pointer.Deref(r.Spec.To)) == 0 {
		verr.AddViolationAt(path, "at least one of 'to' or 'rules' has to be defined")
	}
	topLevel := pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh})
	verr.AddErrorAt(path, validateTo(pointer.Deref(r.Spec.To), topLevel.Kind))
	verr.AddErrorAt(path, validateRules(pointer.Deref(r.Spec.Rules), topLevel.Kind))
	return verr.OrNil()
}

func (r *MeshTimeoutResource) validateTop(targetRef *common_api.TargetRef, isInboundPolicy bool) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	switch core_model.PolicyRole(r.GetMeta()) {
	case mesh_proto.SystemPolicyRole:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.Dataplane,
			},
			GatewayListenerTagsAllowed: true,
			IsInboundPolicy:            isInboundPolicy,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.Dataplane,
			},
			IsInboundPolicy: isInboundPolicy,
		})
	}
}

func validateTo(to []To, topLevelKind common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)

		var supportedKinds []common_api.TargetRefKind
		if topLevelKind == common_api.MeshHTTPRoute {
			supportedKinds = []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshExternalService,
			}
		} else {
			supportedKinds = []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
				common_api.MeshExternalService,
				common_api.MeshMultiZoneService,
				common_api.MeshHTTPRoute,
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

func validateRules(rules []Rule, topLevelKind common_api.TargetRefKind) validators.ValidationError {
	var verr validators.ValidationError
	for idx, rule := range rules {
		path := validators.RootedAt("rules").Index(idx)
		verr.Add(validateMatches(path.Field("matches"), path.Field("default"), pointer.Deref(rule.Matches), rule.Default))
		verr.Add(validateDefault(path.Field("default"), rule.Default, topLevelKind))
	}
	return verr
}

func validateMatches(matchesPath validators.PathBuilder, defaultPath validators.PathBuilder, matches []common_api.Match, conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	hasSpiffeID := false
	for idx, match := range matches {
		matchPath := matchesPath.Index(idx)
		verr.AddErrorAt(matchPath, mesh.ValidateMatch(match))
		if match.SpiffeID == nil && match.SNI == nil {
			verr.AddViolationAt(matchPath, "must specify at least one of 'spiffeID' or 'sni'")
			continue
		}
		if match.SpiffeID != nil {
			hasSpiffeID = true
			switch match.SpiffeID.Type {
			case common_api.ExactMatchType, common_api.PrefixMatchType:
			case "":
				verr.AddViolationAt(matchPath.Field("spiffeID").Field("type"), "must be set")
			default:
				verr.AddViolationAt(matchPath.Field("spiffeID").Field("type"), fmt.Sprintf("unrecognized type %q, supported values are: Exact, Prefix", match.SpiffeID.Type))
			}
		}
	}
	if hasSpiffeID {
		verr.Add(validateSourceConditionedTimeouts(defaultPath, conf))
	}
	return verr
}

func validateSourceConditionedTimeouts(path validators.PathBuilder, conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	const msg = "can't be specified when matches contain spiffeID because this field cannot be conditioned on source identity"

	verr.Add(validators.ValidateNil(path.Field("connectionTimeout"), conf.ConnectionTimeout, msg))
	verr.Add(validators.ValidateNil(path.Field("idleTimeout"), conf.IdleTimeout, msg))

	if http := conf.Http; http != nil {
		httpPath := path.Field("http")
		verr.Add(validators.ValidateNil(httpPath.Field("requestHeadersTimeout"), http.RequestHeadersTimeout, msg))
		verr.Add(validators.ValidateNil(httpPath.Field("maxStreamDuration"), http.MaxStreamDuration, msg))
		verr.Add(validators.ValidateNil(httpPath.Field("maxConnectionDuration"), http.MaxConnectionDuration, msg))
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
