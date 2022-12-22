package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshHealthCheckResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path, validateTo(r.Spec.To))
	return verr.OrNil()
}

func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
		},
	})
	return targetRefErr
}

func validateTo(to []To) validators.ValidationError {
	var verr validators.ValidationError
	for idx, toItem := range to {
		path := validators.RootedAt("to").Index(idx)
		verr.AddErrorAt(path.Field("targetRef"), matcher_validators.ValidateTargetRef(toItem.TargetRef, &matcher_validators.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
		}))

		verr.AddErrorAt(path, validateDefault(toItem.Default))
	}
	return verr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("default")
	verr.Add(validators.ValidateDurationGreaterThanZero(path.Field("interval"), conf.Interval))
	verr.Add(validators.ValidateDurationGreaterThanZero(path.Field("timeout"), conf.Timeout))
	verr.Add(validators.ValidateValueGreaterThanZero(path.Field("unhealthyThreshold"), conf.UnhealthyThreshold))
	verr.Add(validators.ValidateValueGreaterThanZero(path.Field("healthyThreshold"), conf.HealthyThreshold))
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("initialJitter"), conf.InitialJitter))
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("intervalJitter"), conf.IntervalJitter))
	verr.Add(validators.ValidateIntPercentageOrNil(path.Field("intervalJitterPercent"), conf.IntervalJitterPercent))
	verr.Add(validators.ValidateIntPercentageOrNil(path.Field("healthyPanicThreshold"), conf.HealthyPanicThreshold))
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

func validateConfHttp(path validators.PathBuilder, http *HttpHealthCheck) (err validators.ValidationError) {
	err.Add(validators.ValidateStringDefined(path.Field("path"), http.Path))
	err.Add(validateConfHttpExpectedStatuses(path.Field("expectedStatuses"), http.ExpectedStatuses))
	err.Add(validateConfHttpRequestHeadersToAdd(path.Field("requestHeadersToAdd"), http.RequestHeadersToAdd))
	return
}

func validateConfHttpExpectedStatuses(path validators.PathBuilder, expectedStatuses *[]int32) (err validators.ValidationError) {
	if expectedStatuses != nil {
		for i, status := range *expectedStatuses {
			err.Add(validators.ValidateStatusCode(path.Index(i), status))
		}
	}

	return
}

func validateConfHttpRequestHeadersToAdd(path validators.PathBuilder, requestHeadersToAdd *[]HeaderValueOption) (err validators.ValidationError) {
	if requestHeadersToAdd != nil {
		for i, header := range *requestHeadersToAdd {
			path := path.Index(i).Field("header")

			if header.Header == nil {
				err.AddViolationAt(path, validators.MustBeDefined)
				continue
			}

			if header.Header.Key == "" {
				err.AddViolationAt(path.Field("key"), validators.MustNotBeEmpty)
			}
			if header.Header.Value == "" {
				err.AddViolationAt(path.Field("value"), validators.MustNotBeEmpty)
			}
		}
	}

	return
}
