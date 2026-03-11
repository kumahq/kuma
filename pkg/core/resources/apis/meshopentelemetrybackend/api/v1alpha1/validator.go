package v1alpha1

import (
	"math"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

func (r *MeshOpenTelemetryBackendResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	hasEndpoint := r.Spec.Endpoint != nil
	hasNodeEndpoint := r.Spec.NodeEndpoint != nil

	switch {
	case hasEndpoint && hasNodeEndpoint:
		verr.AddViolationAt(path, "exactly one of endpoint or nodeEndpoint must be set, not both")
	case !hasEndpoint && !hasNodeEndpoint:
		verr.AddViolationAt(path, "exactly one of endpoint or nodeEndpoint must be set")
	case hasEndpoint:
		verr.AddErrorAt(path.Field("endpoint"), validateEndpoint(*r.Spec.Endpoint, r.Spec.Protocol))
	case hasNodeEndpoint:
		verr.AddErrorAt(path.Field("nodeEndpoint"), validateNodeEndpoint(*r.Spec.NodeEndpoint, r.Spec.Protocol))
	}

	verr.AddErrorAt(path, validateProtocol(r.Spec.Protocol))
	verr.AddErrorAt(path.Field("env"), validateEnvPolicy(r.Spec.Env))

	return verr.OrNil()
}

func validateEndpoint(endpoint Endpoint, protocol Protocol) validators.ValidationError {
	var verr validators.ValidationError

	if endpoint.Address == "" {
		verr.AddViolationAt(validators.RootedAt("address"), validators.MustNotBeEmpty)
	} else if !govalidator.IsIP(endpoint.Address) && !govalidator.IsDNSName(endpoint.Address) {
		verr.AddViolationAt(validators.RootedAt("address"), "address has to be a valid IP or hostname")
	}

	verr.Add(validatePort(endpoint.Port))
	verr.Add(validatePath(endpoint.Path, protocol))

	return verr
}

func validateNodeEndpoint(endpoint NodeEndpoint, protocol Protocol) validators.ValidationError {
	var verr validators.ValidationError
	verr.Add(validatePort(endpoint.Port))
	verr.Add(validatePath(endpoint.Path, protocol))
	return verr
}

func validatePort(port int32) validators.ValidationError {
	var verr validators.ValidationError
	if port == 0 || port > math.MaxUint16 {
		verr.AddViolationAt(validators.RootedAt("port"), "port must be a valid (1-65535)")
	}
	return verr
}

func validatePath(path *string, protocol Protocol) validators.ValidationError {
	var verr validators.ValidationError
	if path != nil && *path != "" {
		pathField := validators.RootedAt("path")
		if protocol == ProtocolGRPC || protocol == "" {
			verr.AddViolationAt(pathField, "must not be set when protocol is grpc")
			return verr
		}
		if !strings.HasPrefix(*path, "/") {
			verr.AddViolationAt(pathField, "must start with /")
		}
		if strings.ContainsAny(*path, "?#") {
			verr.AddViolationAt(pathField, "must not contain query or fragment")
		}
	}
	return verr
}

func validateProtocol(protocol Protocol) validators.ValidationError {
	var verr validators.ValidationError
	if protocol != "" && protocol != ProtocolGRPC && protocol != ProtocolHTTP {
		verr.AddViolationAt(validators.RootedAt("protocol"), "must be one of: grpc, http")
	}
	return verr
}

func validateEnvPolicy(policy *EnvPolicy) validators.ValidationError {
	var verr validators.ValidationError
	if policy == nil {
		return verr
	}

	if policy.Mode != "" &&
		policy.Mode != EnvModeDisabled &&
		policy.Mode != EnvModeOptional &&
		policy.Mode != EnvModeRequired {
		verr.AddViolationAt(validators.RootedAt("mode"), "must be one of: Disabled, Optional, Required")
	}

	if policy.Precedence != "" &&
		policy.Precedence != EnvPrecedenceExplicitFirst &&
		policy.Precedence != EnvPrecedenceEnvFirst {
		verr.AddViolationAt(validators.RootedAt("precedence"), "must be one of: ExplicitFirst, EnvFirst")
	}

	return verr
}
