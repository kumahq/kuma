package v1alpha1

import (
	"fmt"
	"net/http"
	"strconv"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/validators"
	matcher_validators "github.com/kumahq/kuma/pkg/plugins/policies/matchers/validators"
)

func (r *MeshRetryResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("targetRef"), validateTop(r.Spec.TargetRef))
	if len(r.Spec.To) == 0 {
		verr.AddViolationAt(path.Field("to"), "needs at least one item")
	}
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
				common_api.MeshService,
			},
		}))
		verr.AddErrorAt(path.Field("default"), validateDefault(toItem.Default))
	}
	return verr
}

func validateDefault(conf Conf) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("conf")
	if conf.TCP == nil && conf.HTTP == nil && conf.GRPC == nil {
		verr.AddViolationAt(path, "at least one of the 'tcp', 'http', 'grpc' must be specified")
	}
	if conf.TCP != nil {
		verr.AddErrorAt(path, validateTCP(conf.TCP))
	}
	if conf.HTTP != nil {
		verr.AddErrorAt(path, validateHTTP(conf.HTTP))
	}
	if conf.GRPC != nil {
		verr.AddErrorAt(path, validateGRPC(conf.GRPC))
	}
	return verr
}

func validateTCP(tcp *TCP) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("tcp")
	if tcp.MaxConnectAttempt == nil {
		verr.AddViolationAt(path, validators.MustNotBeEmpty)
	}
	return verr
}

func validateHTTP(http *HTTP) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("http")
	if http.NumRetries == nil && http.PerTryTimeout == nil && http.BackOff == nil && http.RetryOn == nil &&
		http.RetriableRequestHeaders == nil && http.RetriableResponseHeaders == nil && http.RateLimitedBackOff == nil {
		verr.AddViolationAt(path, validators.MustNotBeEmpty)
	}
	if http.BackOff != nil {
		verr.AddErrorAt(path, validateBackOff(http.BackOff))
	}
	if http.RateLimitedBackOff != nil {
		verr.AddErrorAt(path, validateRateLimitedBackOff(http.RateLimitedBackOff))
	}
	if http.RetryOn != nil {
		verr.AddErrorAt(path, validateHTTPRetryOn(*http.RetryOn))
	}
	return verr
}

func validateHTTPRetryOn(retryOn []HTTPRetryOn) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("retryOn")
	for idx, ro := range retryOn {
		switch ro {
		case All5xx:
		case GatewayError:
		case Reset:
		case Retriable4xx:
		case ConnectFailure:
		case EnvoyRatelimited:
		case RefusedStream:
		case Http3PostConnectFailure:
		case HttpMethodConnect:
		case HttpMethodDelete:
		case HttpMethodGet:
		case HttpMethodHead:
		case HttpMethodOptions:
		case HttpMethodPatch:
		case HttpMethodPost:
		case HttpMethodPut:
		case HttpMethodTrace:
		default:
			// method http.StatusText returns empty string for unknown status codes
			if i, err := strconv.Atoi(string(ro)); err != nil || http.StatusText(i) == "" {
				verr.AddViolationAt(path.Index(idx), fmt.Sprintf("unknown item '%v'", ro))
				continue
			}
		}
	}
	return verr
}

func validateGRPC(grpc *GRPC) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("grpc")
	if grpc.NumRetries == nil && grpc.PerTryTimeout == nil && grpc.BackOff == nil && grpc.RetryOn == nil && grpc.RateLimitedBackOff == nil {
		verr.AddViolationAt(path, validators.MustNotBeEmpty)
	}
	if grpc.BackOff != nil {
		verr.AddErrorAt(path, validateBackOff(grpc.BackOff))
	}
	if grpc.RateLimitedBackOff != nil {
		verr.AddErrorAt(path, validateRateLimitedBackOff(grpc.RateLimitedBackOff))
	}
	if grpc.RetryOn != nil {
		verr.AddErrorAt(path, validateGRPCRetryOn(*grpc.RetryOn))
	}
	return verr
}

func validateBackOff(b *BackOff) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("backOff")
	if b.BaseInterval == nil && b.MaxInterval == nil {
		verr.AddViolationAt(path, validators.MustNotBeEmpty)
	}
	return verr
}

func validateRateLimitedBackOff(rateLimitedBackOff *RateLimitedBackOff) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("rateLimitedBackOff")

	if rateLimitedBackOff.MaxInterval != nil {
		verr.Add(validators.ValidateDurationGreaterThanZero(path.Field("maxInterval"), *rateLimitedBackOff.MaxInterval))
	}

	if len(rateLimitedBackOff.ResetHeaders) == 0 {
		verr.AddViolationAt(path.Field("resetHeaders"), validators.MustBeDefined)
	}

	for i, header := range rateLimitedBackOff.ResetHeaders {
		index := path.Field("resetHeaders").Index(i)
		verr.Add(validators.ValidateStringDefined(index.Field("name"), header.Name))
	}

	return verr
}

func validateGRPCRetryOn(retryOn []GRPCRetryOn) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("retryOn")
	for idx, ro := range retryOn {
		switch ro {
		case Canceled:
		case DeadlineExceeded:
		case Internal:
		case ResourceExhausted:
		case Unavailable:
		default:
			verr.AddViolationAt(path.Index(idx), fmt.Sprintf("unknown item '%v'", ro))
		}
	}
	return verr
}
