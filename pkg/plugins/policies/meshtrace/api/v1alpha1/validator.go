package v1alpha1

import (
	"fmt"
	"math"
	net_url "net/url"
	"strconv"

	"github.com/asaskevich/govalidator"
	"github.com/shopspring/decimal"
	"k8s.io/apimachinery/pkg/util/intstr"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

func (r *MeshTraceResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), r.validateTop(r.Spec.TargetRef, r.Descriptor()))
	verr.AddErrorAt(path.Field("default"), validateDefault(r.Spec.Default))
	return verr.OrNil()
}

func (r *MeshTraceResource) validateTop(targetRef *common_api.TargetRef, descriptor core_model.ResourceTypeDescriptor) validators.ValidationError {
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
			},
			GatewayListenerTagsAllowed: false,
			Descriptor:                 descriptor,
		})
	default:
		return mesh.ValidateTargetRef(*targetRef, &mesh.ValidateTargetRefOpts{
			SupportedKinds: []common_api.TargetRefKind{
				common_api.Mesh,
				common_api.MeshSubset,
				common_api.MeshService,
				common_api.MeshServiceSubset,
			},
			Descriptor: descriptor,
		})
	}
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError

	backendsPath := validators.RootedAt("backends")
	if conf.Backends == nil {
		verr.AddViolationAt(backendsPath, validators.MustBeDefined)
	}

	switch len(pointer.Deref(conf.Backends)) {
	case 0:
		break
	case 1:
		verr.AddError("", validateBackend(conf, backendsPath))
	default:
		verr.AddViolationAt(backendsPath, "must have zero or one backend defined")
	}

	for tagIndex, tag := range pointer.Deref(conf.Tags) {
		path := validators.RootedAt("tags").Index(tagIndex)
		if tag.Name == "" {
			verr.AddViolationAt(path.Field("name"), validators.MustNotBeEmpty)
		}

		if (tag.Header != nil) == (tag.Literal != nil) {
			verr.AddViolationAt(path, validators.MustHaveOnlyOne("tag", "header", "literal"))
		}
	}

	if conf.Sampling != nil {
		if client := conf.Sampling.Client; client != nil {
			verr.AddErrorAt(validators.RootedAt("sampling").Field("client"), validateSampling(*client))
		}
		if random := conf.Sampling.Random; random != nil {
			verr.AddErrorAt(validators.RootedAt("sampling").Field("random"), validateSampling(*random))
		}
		if overall := conf.Sampling.Overall; overall != nil {
			verr.AddErrorAt(validators.RootedAt("sampling").Field("overall"), validateSampling(*overall))
		}
	}

	return verr
}

func validateBackend(conf Conf, backendsPath validators.PathBuilder) validators.ValidationError {
	var verr validators.ValidationError
	backend := pointer.Deref(conf.Backends)[0]
	firstBackendPath := backendsPath.Index(0)

	switch backend.Type {
	case DatadogBackendType:
		datadogPath := firstBackendPath.Field("datadog")
		datadogBackend := backend.Datadog
		if datadogBackend == nil {
			verr.AddViolationAt(datadogPath, validators.MustBeDefined)
			break
		}

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
	case ZipkinBackendType:
		zipkinPath := firstBackendPath.Field("zipkin")
		zipkinBackend := backend.Zipkin
		if zipkinBackend == nil {
			verr.AddViolationAt(zipkinPath, validators.MustBeDefined)
			break
		}

		if zipkinBackend.Url == "" {
			verr.AddViolationAt(zipkinPath.Field("url"), validators.MustNotBeEmpty)
		} else if !govalidator.IsURL(zipkinBackend.Url) {
			verr.AddViolationAt(zipkinPath.Field("url"), "must be a valid url")
		}
	case OpenTelemetryBackendType:
		otelPath := firstBackendPath.Field("openTelemetry")
		otelBackend := backend.OpenTelemetry
		if otelBackend == nil {
			verr.AddViolationAt(otelPath, validators.MustBeDefined)
			break
		}
	default:
		panic(fmt.Sprintf("unknown backend type %v", backend.Type))
	}

	return verr
}

func validateSampling(sampling intstr.IntOrString) validators.ValidationError {
	var verr validators.ValidationError

	dec, err := common_api.NewDecimalFromIntOrString(sampling)
	if err != nil {
		verr.AddViolation("", "string is not a number")
	}

	if dec.LessThan(decimal.Zero) || dec.GreaterThan(decimal.NewFromInt(100)) {
		verr.AddViolation("", "must be between 0 and 100")
	}

	return verr
}
