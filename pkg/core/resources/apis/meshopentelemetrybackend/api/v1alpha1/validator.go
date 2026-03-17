package v1alpha1

import (
	"slices"
	"strings"

	"github.com/asaskevich/govalidator"

	"github.com/kumahq/kuma/v2/pkg/core/validators"
)

func (r *MeshOpenTelemetryBackendResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	protocol := ProtocolGRPC
	if r.Spec.Protocol != nil {
		protocol = *r.Spec.Protocol
	}

	if r.Spec.Endpoint != nil {
		verr.AddErrorAt(path.Field("endpoint"), validateEndpoint(*r.Spec.Endpoint, protocol))
	}

	verr.AddErrorAt(path, validateProtocol(protocol))
	verr.AddErrorAt(path.Field("env"), validateEnvPolicy(r.Spec.Env))

	return verr.OrNil()
}

func validateEndpoint(endpoint Endpoint, protocol Protocol) validators.ValidationError {
	var verr validators.ValidationError

	if endpoint.Address != nil {
		if *endpoint.Address == "" {
			verr.AddViolationAt(validators.RootedAt("address"), validators.MustNotBeEmpty)
		} else if !govalidator.IsIP(*endpoint.Address) && !govalidator.IsDNSName(*endpoint.Address) {
			verr.AddViolationAt(validators.RootedAt("address"), "address has to be a valid IP or hostname")
		}
	}

	if endpoint.Port != nil {
		verr.Add(validators.ValidatePort(validators.RootedAt("port"), uint32(*endpoint.Port)))
	}
	verr.Add(validatePath(endpoint.Path, protocol))

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

var allProtocols = []string{string(ProtocolGRPC), string(ProtocolHTTP)}

func validateProtocol(protocol Protocol) validators.ValidationError {
	var verr validators.ValidationError
	if protocol != "" && !slices.Contains(allProtocols, string(protocol)) {
		verr.AddErrorAt(validators.RootedAt("protocol"), validators.MakeFieldMustBeOneOfErr("protocol", allProtocols...))
	}
	return verr
}

var (
	allEnvModes       = []string{string(EnvModeDisabled), string(EnvModeOptional), string(EnvModeRequired)}
	allEnvPrecedences = []string{string(EnvPrecedenceExplicitFirst), string(EnvPrecedenceEnvFirst)}
)

func validateEnvPolicy(policy *EnvPolicy) validators.ValidationError {
	var verr validators.ValidationError
	if policy == nil {
		return verr
	}

	if policy.Mode != "" && !slices.Contains(allEnvModes, string(policy.Mode)) {
		verr.AddErrorAt(validators.RootedAt("mode"), validators.MakeFieldMustBeOneOfErr("mode", allEnvModes...))
	}

	if policy.Precedence != "" && !slices.Contains(allEnvPrecedences, string(policy.Precedence)) {
		verr.AddErrorAt(validators.RootedAt("precedence"), validators.MakeFieldMustBeOneOfErr("precedence", allEnvPrecedences...))
	}

	return verr
}
