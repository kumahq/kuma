package v1alpha1

import (
	"fmt"
	"math"
	"strings"

	"github.com/asaskevich/govalidator"
	"golang.org/x/exp/slices"

	common_proto "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
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
		verr.AddViolation("", validators.MustBeDefined)
		return verr
	}

	backendsPath := validators.RootedAt("backends")

	switch len(conf.GetBackends()) {
	case 0:
		break
	case 1:
		verr.AddError("", validateBackend(conf, backendsPath))
	default:
		verr.AddViolationAt(backendsPath, "must have zero or one backend defined")
	}

	tags := conf.GetTags()
	for tagIndex, tag := range tags {
		path := validators.RootedAt("tags").Index(tagIndex)
		if tag.GetName() == "" {
			verr.AddViolationAt(path.Field("name"), validators.MustNotBeEmpty)
		}

		if (tag.GetHeader() != nil) == (tag.GetLiteral() != "") {
			verr.AddViolationAt(path, validators.MustHaveOnlyOne("tag", "header", "literal"))
		}
	}

	sampling := conf.GetSampling()
	if sampling != nil {
		if sampling.GetClient() != nil {
			verr.AddErrorAt(validators.RootedAt("sampling").Field("client"), validateSampling(sampling.GetClient().GetValue()))
		}
		if sampling.GetRandom() != nil {
			verr.AddErrorAt(validators.RootedAt("sampling").Field("random"), validateSampling(sampling.GetRandom().GetValue()))
		}
		if sampling.GetOverall() != nil {
			verr.AddErrorAt(validators.RootedAt("sampling").Field("overall"), validateSampling(sampling.GetOverall().GetValue()))
		}
	}

	return verr
}

func validateBackend(conf *MeshTrace_Conf, backendsPath validators.PathBuilder) validators.ValidationError {
	var verr validators.ValidationError
	backend := conf.GetBackends()[0]
	firstBackendPath := backendsPath.Index(0)
	if (backend.GetDatadog() != nil) == (backend.GetZipkin() != nil) {
		verr.AddViolationAt(firstBackendPath, validators.MustHaveOnlyOne("backend", "datadog", "zipkin"))
	}

	if backend.GetDatadog() != nil {
		datadogBackend := backend.GetDatadog()
		datadogPath := firstBackendPath.Field("datadog")
		if datadogBackend.Address == "" {
			verr.AddViolationAt(datadogPath.Field("address"), "must not be empty")
		} else if !govalidator.IsURL(datadogBackend.Address) {
			verr.AddViolationAt(datadogPath.Field("address"), "must be a valid address")
		}

		if datadogBackend.Port == 0 || datadogBackend.Port > math.MaxUint16 {
			verr.AddViolationAt(datadogPath.Field("port"), fmt.Sprintf("must be a valid port (1-%d)", math.MaxUint16))
		}
	}

	if backend.GetZipkin() != nil {
		zipkinBackend := backend.GetZipkin()
		zipkinPath := firstBackendPath.Field("zipkin")

		if zipkinBackend.Url == "" {
			verr.AddViolationAt(zipkinPath.Field("url"), validators.MustNotBeEmpty)
		} else if !govalidator.IsURL(zipkinBackend.Url) {
			verr.AddViolationAt(zipkinPath.Field("url"), "must be a valid url")
		}

		if zipkinBackend.ApiVersion != "" {
			validZipkinApiVersions := []string{"httpJson", "httpProto"}
			if !slices.Contains(validZipkinApiVersions, zipkinBackend.ApiVersion) {
				verr.AddViolationAt(zipkinPath.Field("apiVersion"), fmt.Sprintf("must be one of %s", strings.Join(validZipkinApiVersions, ", ")))
			}
		}
	}

	return verr
}

func validateSampling(sampling uint32) validators.ValidationError {
	var verr validators.ValidationError

	if sampling > 100 {
		verr.AddViolation("", "must be between 0 and 100")
	}

	return verr
}
