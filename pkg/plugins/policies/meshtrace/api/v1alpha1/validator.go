package v1alpha1

import (
	"fmt"
	"math"
	net_url "net/url"
	"strconv"
	"strings"

	"github.com/asaskevich/govalidator"
	"golang.org/x/exp/slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshTraceResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}
func validateTop(targetRef common_api.TargetRef) validators.ValidationError {
	targetRefErr := matcher_validators.ValidateTargetRef(targetRef, &matcher_validators.ValidateTargetRefOpts{
		SupportedKinds: []common_api.TargetRefKind{
			common_api.Mesh,
			common_api.MeshSubset,
			common_api.MeshService,
			common_api.MeshServiceSubset,
			common_api.MeshGatewayRoute,
		},
	})
	return targetRefErr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError

	backendsPath := validators.RootedAt("backends")
	if conf.Backends == nil {
		verr.AddViolationAt(backendsPath, validators.MustBeDefined)
	}

	switch len(conf.Backends) {
	case 0:
		break
	case 1:
		verr.AddError("", validateBackend(conf, backendsPath))
	default:
		verr.AddViolationAt(backendsPath, "must have zero or one backend defined")
	}

	for tagIndex, tag := range conf.Tags {
		path := validators.RootedAt("tags").Index(tagIndex)
		if tag.Name == "" {
			verr.AddViolationAt(path.Field("name"), validators.MustNotBeEmpty)
		}

		if (tag.Header != nil) == (tag.Literal != "") {
			verr.AddViolationAt(path, validators.MustHaveOnlyOne("tag", "header", "literal"))
		}
	}

	if client := conf.Sampling.Client; client != nil {
		verr.AddErrorAt(validators.RootedAt("sampling").Field("client"), validateSampling(*client))
	}
	if random := conf.Sampling.Random; random != nil {
		verr.AddErrorAt(validators.RootedAt("sampling").Field("random"), validateSampling(*random))
	}
	if overall := conf.Sampling.Overall; overall != nil {
		verr.AddErrorAt(validators.RootedAt("sampling").Field("overall"), validateSampling(*overall))
	}

	return verr
}

func validateBackend(conf Conf, backendsPath validators.PathBuilder) validators.ValidationError {
	var verr validators.ValidationError
	backend := conf.Backends[0]
	firstBackendPath := backendsPath.Index(0)
	if (backend.Datadog != nil) == (backend.Zipkin != nil) {
		verr.AddViolationAt(firstBackendPath, validators.MustHaveOnlyOne("backend", "datadog", "zipkin"))
	}

	if backend.Datadog != nil {
		datadogBackend := backend.Datadog
		datadogPath := firstBackendPath.Field("datadog")

		url, err := net_url.ParseRequestURI(datadogBackend.Url)
		if err != nil {
			verr.AddViolationAt(datadogPath.Field("url"), "must be a valid url")
		} else {
			// taken from https://github.com/DataDog/dd-trace-go/blob/acd5c8b03e186971808ddd0a42b89b4399068345/profiler/options.go#L312
			if url.Scheme != "http" {
				verr.AddViolationAt(datadogPath.Field("url"), "scheme must be http")
			}
			if url.Port() == "" {
				verr.AddViolationAt(datadogPath.Field("url"), "port "+validators.MustBeDefined)
			} else {
				port, err := strconv.Atoi(url.Port())
				if err != nil {
					verr.AddViolationAt(datadogPath.Field("url"), "port must be a valid (1-65535)")
				} else if port == 0 || port > math.MaxUint16 {
					verr.AddViolationAt(datadogPath.Field("url"), "port must be a valid (1-65535)")
				}
			}

			otherFieldsEmpty := map[string]bool{
				"opaque":   url.Opaque == "",
				"user":     url.User == nil,
				"path":     url.Path == "",
				"query":    url.RawQuery == "",
				"fragment": url.Fragment == "",
			}
			var nonEmptyFields []string
			for name, empty := range otherFieldsEmpty {
				if !empty {
					nonEmptyFields = append(nonEmptyFields, name)
				}
			}

			for _, nonEmptyField := range nonEmptyFields {
				verr.AddViolationAt(datadogPath.Field("url"), nonEmptyField+" "+validators.MustNotBeDefined)
			}
		}
	}

	if backend.Zipkin != nil {
		zipkinBackend := backend.Zipkin
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
