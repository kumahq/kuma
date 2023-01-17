package v1alpha1

import (
	"time"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshRateLimitResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	if len(r.Spec.From) == 0 {
		verr.AddViolationAt(path.Field("from"), "needs at least one item")
	}
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef, hasTcpConfiguration(r.Spec.From)))
	verr.AddErrorAt(path, validateFrom(r.Spec.From))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef, hasTcpConfig bool) validators.ValidationError {
	supportedKinds := []common_api.TargetRefKind{
		common_api.Mesh,
		common_api.MeshSubset,
		common_api.MeshService,
		common_api.MeshServiceSubset,
	}
	if !hasTcpConfig {
		supportedKinds = append(supportedKinds, common_api.MeshGatewayRoute)
	}
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: supportedKinds,
	})
	return targetRefErr
}

func validateFrom(from []From) validators.ValidationError {
	var verr validators.ValidationError
	for idx, fromItem := range from {
		path := validators.RootedAt("from").Index(idx)
		defaultField := path.Field("default")
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(fromItem.GetTargetRef(), &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
			},
		}))
		verr.Add(validateDefault(defaultField, fromItem.Default))
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

func hasTcpConfiguration(from []From) bool {
	for _, fromItem := range from {
		if isTcp(fromItem.Default) {
			return true
		}
	}
	return false
}

func isTcp(conf Conf) bool {
	return pointer.Deref(conf.Local).TCP != nil
}
