package v1alpha1

import (
	"github.com/shopspring/decimal"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshCircuitBreakerResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	if len(r.Spec.To) == 0 && len(r.Spec.From) == 0 {
		verr.AddViolationAt(path, "at least one of 'from', 'to' has to be defined")
	}
	verr.AddErrorAt(path, validateFrom(r.Spec.From))
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
			common_api.MeshService,
			common_api.MeshGateway,
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
		verr.Add(validateDefault(defaultField, fromItem.Default))
	}
	return verr
}

func validateTo(topTargetRef common_api.TargetRef, to []To) validators.ValidationError {
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
		if toItem.TargetRef.Kind == common_api.MeshExternalService && topTargetRef.Kind != common_api.Mesh {
			verr.AddViolationAt(path.Field("targetRef.kind"), "kind MeshExternalService is only allowed with targetRef.kind: Mesh as it is configured on the Zone Egress and shared by all clients in the mesh")
		}
		defaultField := path.Field("default")
		verr.Add(validateDefault(defaultField, toItem.Default))
	}
	return verr
}

func validateDefault(path validators.PathBuilder, conf Conf) validators.ValidationError {
	var verr validators.ValidationError

	if conf.ConnectionLimits == nil && conf.OutlierDetection == nil {
		verr.AddViolationAt(path, "at least one of: 'connectionLimits' or 'outlierDetection' should be configured")
		return verr
	}

	verr.Add(validateConnectionLimits(path.Field("connectionLimits"), conf.ConnectionLimits))
	verr.Add(validateOutlierDetection(path.Field("outlierDetection"), conf.OutlierDetection))
	return verr
}

func validateConnectionLimits(path validators.PathBuilder, limits *ConnectionLimits) validators.ValidationError {
	var verr validators.ValidationError
	if limits == nil {
		return verr
	}

	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("maxConnections"), limits.MaxConnections))
	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("maxConnectionPools"), limits.MaxConnectionPools))
	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("maxPendingRequests"), limits.MaxPendingRequests))
	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("maxRetries"), limits.MaxRetries))
	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("maxRequests"), limits.MaxRequests))

	return verr
}

func validateOutlierDetection(path validators.PathBuilder, outlierDetection *OutlierDetection) validators.ValidationError {
	var verr validators.ValidationError
	if outlierDetection == nil {
		return verr
	}

	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("interval"), outlierDetection.Interval))
	verr.Add(validators.ValidateDurationGreaterThanZeroOrNil(path.Field("baseEjectionTime"), outlierDetection.BaseEjectionTime))
	verr.Add(validators.ValidateUInt32PercentageOrNil(path.Field("maxEjectionPercent"), outlierDetection.MaxEjectionPercent))

	verr.Add(validateDetectors(path.Field("detectors"), outlierDetection.Detectors))

	return verr
}

func validateDetectors(path validators.PathBuilder, detectors *Detectors) validators.ValidationError {
	var verr validators.ValidationError
	if detectors == nil {
		verr.AddViolationAt(path, validators.MustBeDefined)
		return verr
	}

	if detectors.FailurePercentage == nil && detectors.GatewayFailures == nil &&
		detectors.LocalOriginFailures == nil && detectors.TotalFailures == nil &&
		detectors.SuccessRate == nil {
		verr.AddViolationAt(path, validators.MustHaveAtLeastOne("totalFailures", "gatewayFailures", "localOriginFailures", "successRate", "failurePercentage"))
		return verr
	}

	verr.Add(validateDetectorTotalFailures(path.Field("totalFailures"), detectors.TotalFailures))
	verr.Add(validateDetectorGatewayFailures(path.Field("gatewayFailures"), detectors.GatewayFailures))
	verr.Add(validateDetectorLocalOriginFailures(path.Field("localOriginFailures"), detectors.LocalOriginFailures))
	verr.Add(validateDetectorSuccessRate(path.Field("successRate"), detectors.SuccessRate))
	verr.Add(validateDetectorFailurePercentage(path.Field("failurePercentage"), detectors.FailurePercentage))

	return verr
}

func validateDetectorTotalFailures(path validators.PathBuilder, detector *DetectorTotalFailures) validators.ValidationError {
	var verr validators.ValidationError
	if detector == nil {
		return verr
	}

	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("consecutive"), detector.Consecutive))

	return verr
}

func validateDetectorGatewayFailures(path validators.PathBuilder, detector *DetectorGatewayFailures) validators.ValidationError {
	var verr validators.ValidationError
	if detector == nil {
		return verr
	}

	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("consecutive"), detector.Consecutive))

	return verr
}

func validateDetectorLocalOriginFailures(path validators.PathBuilder, detector *DetectorLocalOriginFailures) validators.ValidationError {
	var verr validators.ValidationError
	if detector == nil {
		return verr
	}

	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("consecutive"), detector.Consecutive))

	return verr
}

func validateDetectorSuccessRate(path validators.PathBuilder, detector *DetectorSuccessRateFailures) validators.ValidationError {
	var verr validators.ValidationError
	if detector == nil {
		return verr
	}

	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("minimumHosts"), detector.MinimumHosts))
	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("requestVolume"), detector.RequestVolume))
	if detector.StandardDeviationFactor != nil {
		dec, err := common_api.NewDecimalFromIntOrString(*detector.StandardDeviationFactor)
		if err != nil {
			verr.AddViolationAt(path.Field("standardDeviationFactor"), "invalid number")
		} else if dec.LessThanOrEqual(decimal.Zero) {
			verr.AddViolationAt(path.Field("standardDeviationFactor"), validators.HasToBeGreaterThanZero)
		}
	}

	return verr
}

func validateDetectorFailurePercentage(path validators.PathBuilder, detector *DetectorFailurePercentageFailures) validators.ValidationError {
	var verr validators.ValidationError
	if detector == nil {
		return verr
	}

	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("minimumHosts"), detector.MinimumHosts))
	verr.Add(validators.ValidateIntegerGreaterThanZeroOrNil(path.Field("requestVolume"), detector.RequestVolume))
	verr.Add(validators.ValidateUInt32PercentageOrNil(path.Field("threshold"), detector.Threshold))

	return verr
}
