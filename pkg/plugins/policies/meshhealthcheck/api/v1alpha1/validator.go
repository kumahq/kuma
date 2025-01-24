package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshHealthCheckResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path, validateTo(pointer.DerefOr(r.Spec.TargetRef, common_api.TargetRef{Kind: common_api.Mesh}), r.Spec.To))
	return verr.OrNil()
}

func validateTop(targetRef *common_api.TargetRef) validators.ValidationError {
	if targetRef == nil {
		return validators.ValidationError{}
	}
	targetRefErr := mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshGateway,
			common_api.MeshService,
			common_api.MeshServiceSubset,
			common_api.Dataplane,
		},
		GatewayListenerTagsAllowed: true,
	})
	return targetRefErr
}

func validateTo(topTargetRef common_api.TargetRef, to []To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), mesh.ValidateTargetRef(toItem.TargetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshService,
				common_api.MeshMultiZoneService,
			},
		}))
		if toItem.TargetRef.Kind == common_api.MeshExternalService && topTargetRef.Kind != common_api.Mesh {
			verr.AddViolationAt(path.Field("targetRef.kind"), "kind MeshExternalService is only allowed with targetRef.kind: Mesh as it is configured on the Zone Egress and shared by all clients in the mesh")
		}
		verr.AddErrorAt(path, validateDefault(toItem.Default))
	}
	return verr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("default")
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("interval"), conf.Interval))
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("timeout"), conf.Timeout))
	verr.Add(validators.ValidateValueGreaterThanZeroOrNil(path.Field("unhealthyThreshold"), conf.UnhealthyThreshold))
	verr.Add(validators.ValidateValueGreaterThanZeroOrNil(path.Field("healthyThreshold"), conf.HealthyThreshold))
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("initialJitter"), conf.InitialJitter))
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("intervalJitter"), conf.IntervalJitter))
	verr.Add(validators.ValidateIntPercentageOrNil(path.Field("intervalJitterPercent"), conf.IntervalJitterPercent))
	verr.Add(validators.ValidatePercentageOrNil(path.Field("healthyPanicThreshold"), conf.HealthyPanicThreshold))
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("noTrafficInterval"), conf.NoTrafficInterval))
	verr.Add(validators.ValidatePathOrNil(path.Field("eventLogPath"), conf.EventLogPath))
	if conf.Http != nil {
		verr.Add(validateConfHttp(path.Field("http"), conf.Http))
	}
	// there is nothing to check in tcp & gRPC because all fields are optional
	if conf.Http == nil && conf.Tcp == nil && conf.Grpc == nil {
		verr.AddViolationAt(path, validators.MustHaveAtLeastOne("http", "tcp", "grpc"))
	}
	return verr
}

func validateConfHttp(path validators.PathBuilder, http *HttpHealthCheck) validators.ValidationError {
	var err validators.ValidationError
	if http.Path != nil {
		err.Add(validators.ValidateStringDefined(path.Field("path"), *http.Path))
	}
	err.Add(validateConfHttpExpectedStatuses(path.Field("expectedStatuses"), http.ExpectedStatuses))
	return err
}

func validateConfHttpExpectedStatuses(path validators.PathBuilder, expectedStatuses *[]int32) validators.ValidationError {
	var err validators.ValidationError
	if expectedStatuses != nil {
		for i, status := range *expectedStatuses {
			err.Add(validators.ValidateStatusCode(path.Index(i), status))
		}
	}

	return err
}
