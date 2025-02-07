package v1alpha1

import (
	"fmt"
	"time"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/inbound"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshRateLimitResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	if len(r.Spec.Rules) > 0 && (len(r.Spec.To) > 0 || len(r.Spec.From) > 0) {
		verr.AddViolationAt(path, "fields 'to' and 'from' must be empty when 'rules' is defined")
	}
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef, inbound.AffectsInbounds(r.Spec)))
	topLevel := pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh})
	verr.AddErrorAt(path, validateRules(topLevel, r.Spec.Rules))
	verr.AddErrorAt(path, validateFrom(topLevel, r.Spec.From))
	verr.AddErrorAt(path, validateTo(topLevel, r.Spec.To))
	return verr.OrNil()
}

func (r *MeshRateLimitResource) validateTop(targetRef *common_api.TargetRef, isInboundPolicy bool) validators.ValidationError {
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
				common_api.Dataplane,
			},
			GatewayListenerTagsAllowed: true,
			IsInboundPolicy:            isInboundPolicy,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
				common_api.Dataplane,
			},
			IsInboundPolicy: isInboundPolicy,
		})
	}
}

func validateRules(topTargetRef common_api.TargetRef, rules []Rule) validators.ValidationError {
	var verr validators.ValidationError
	if common_api.IncludesGateways(topTargetRef) && len(rules) != 0 {
		verr.AddViolationAt(validators.RootedAt("rules"), validators.MustNotBeDefined)
		return verr
	}
	if topTargetRef.Kind == common_api.MeshHTTPRoute && len(rules) != 0 {
		verr.AddViolationAt(validators.RootedAt("rules"), validators.MustNotBeDefined)
		return verr
	}
	for idx, ruleItem := range rules {
		path := validators.RootedAt("rules").Index(idx)
		verr.Add(validateDefault(path.Field("default"), ruleItem.Default))
	}
	return verr
}

func validateFrom(topTargetRef common_api.TargetRef, from []From) validators.ValidationError {
	var verr validators.ValidationError
	if common_api.IncludesGateways(topTargetRef) && len(from) != 0 {
		verr.AddViolationAt(validators.RootedAt("from"),
			fmt.Sprintf("%s when the scope includes a Gateway, select only proxyType Sidecar or select only gateways and use spec.to", validators.MustNotBeDefined))
		return verr
	}
	if topTargetRef.Kind == common_api.MeshHTTPRoute && len(from) != 0 {
		verr.AddViolationAt(validators.RootedAt("from"),
			fmt.Sprintf("%s when spec.kind is MeshHTTPRoute", validators.MustNotBeDefined))
		return verr
	}
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		defaultField := path.Field("default")
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(fromItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		}))
		verr.Add(validateDefault(defaultField, fromItem.Default))
	}
	return verr
}

func validateTo(topTargetRef common_api.TargetRef, to []To) validators.ValidationError {
	var verr validators.ValidationError
	if !common_api.IncludesGateways(topTargetRef) && len(to) != 0 {
		verr.AddViolationAt(validators.RootedAt("to"), validators.MustNotBeDefined)
		return verr
	}
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		defaultField := path.Field("default")
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(toItem.GetTargetRef(), &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		}))
		verr.Add(validateDefault(defaultField, toItem.Default))
	}
	return verr
}

func validateDefault(path validators.PathBuilder, conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	path = path.Field("local")

	if conf.Local == nil {
		verr.AddViolationAt(path, validators.MustBeDefined)
		return verr
	}

	if conf.Local.TCP == nil && conf.Local.HTTP == nil {
		verr.AddViolationAt(path, validators.MustHaveAtLeastOne("tcp", "http"))
	}

	if conf.Local.HTTP != nil {
		verr.Add(validateLocalHttp(path.Field("http"), conf.Local.HTTP))
	}

	if conf.Local.TCP != nil {
		verr.Add(validateLocalTcp(path.Field("tcp"), conf.Local.TCP))
	}

	return verr
}

func validateLocalHttp(path validators.PathBuilder, localHttp *LocalHTTP) validators.ValidationError {
	var verr validators.ValidationError
	if localHttp.Disabled == nil && localHttp.RequestRate == nil && localHttp.OnRateLimit == nil {
		verr.AddViolationAt(path, validators.MustHaveAtLeastOne("disabled", "requestRate", "onRateLimit"))
		return verr
	}
	if localHttp.RequestRate != nil {
		verr.Add(validateRate(path.Field("requestRate"), localHttp.RequestRate))
	}
	if localHttp.OnRateLimit != nil {
		path = path.Field("onRateLimit")
		verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("status"), localHttp.OnRateLimit.Status))
	}
	return verr
}

func validateLocalTcp(path validators.PathBuilder, localTcp *LocalTCP) validators.ValidationError {
	var verr validators.ValidationError
	if localTcp.Disabled == nil && localTcp.ConnectionRate == nil {
		verr.AddViolationAt(path, validators.MustHaveAtLeastOne("disabled", "connectionRate"))
		return verr
	}
	if localTcp.ConnectionRate != nil {
		verr.Add(validateRate(path.Field("connectionRate"), localTcp.ConnectionRate))
	}
	return verr
}

func validateRate(path validators.PathBuilder, rate *Rate) validators.ValidationError {
	var verr validators.ValidationError
	verr.Add(validators.ValidateIntegerGreaterThan(path.Field("num"), rate.Num, 0))
	verr.Add(validators.ValidateDurationGreaterThan(path.Field("interval"), &rate.Interval, 50*time.Millisecond))
	return verr
}
