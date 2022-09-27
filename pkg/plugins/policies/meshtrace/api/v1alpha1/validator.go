package v1alpha1

import (
    common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
    "github.com/kumahq/kuma/pkg/core/validators"
    matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
    "github.com/kumahq/kuma/pkg/util/validation"
)

func (r *MeshTraceResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.GetTargetRef()))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.GetDefault()))
	return verr.OrNil()
}
func validateTop(targetRef *common_proto.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_proto.TargetRef_Kind{
			common_proto.TargetRef_Mesh,
			common_proto.TargetRef_MeshSubset,
			common_proto.TargetRef_MeshService,
			common_proto.TargetRef_MeshServiceSubset,
			common_proto.TargetRef_MeshGatewayRoute,
		},
	})
	return targetRefErr
}

func validateDefault(conf *MeshTrace_Conf) validators.ValidationError {
	var verr validators.ValidationError

	if conf == nil {
		verr.AddViolation("", validation.MustBeDefined())
		return verr
	}

	if len(conf.GetBackends()) != 1 {
		verr.AddViolation("backends", "must have only one backend")
	} else {
		backend := conf.GetBackends()[0]
		datadog := validation.Bool2Int(backend.GetDatadog() != nil)
		zipkin := validation.Bool2Int(backend.GetZipkin() != nil)

		if datadog + zipkin != 1 {
			verr.AddViolation("backend", validation.MustHaveOnlyOne("backend[0]", "datadog", "zipkin"))
		}
	}

	tags := conf.GetTags()
	for tagIndex, tag := range tags {
		indexedField := validators.RootedAt("tags").Index(tagIndex)
		if tag.GetName() == "" {
			verr.AddViolationAt(indexedField, "tag's name must not be empty")
		}

		header := validation.Bool2Int(tag.GetHeader() != nil)
		literal := validation.Bool2Int(tag.GetLiteral() != "")

		if header + literal != 1 {
			verr.AddViolationAt(indexedField, validation.MustHaveOnlyOne("tag", "header", "literal"))
		}
	}

	sampling := conf.GetSampling()
	if sampling != nil {
		verr.AddErrorAt(validators.RootedAt("sampling").Field("client"), validateSampling(sampling.GetClient()))
		verr.AddErrorAt(validators.RootedAt("sampling").Field("random"), validateSampling(sampling.GetRandom()))
		verr.AddErrorAt(validators.RootedAt("sampling").Field("overall"), validateSampling(sampling.GetOverall()))
	}

	return verr
}

func validateSampling(sampling float64) validators.ValidationError {
	var verr validators.ValidationError

	if sampling < 0 || sampling > 100 {
		verr.AddViolation("", "must be between 0 and 100")
	}

	return verr
}
