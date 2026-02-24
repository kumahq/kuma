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

	verr.AddErrorAt(path.Field("endpoint"), validateEndpoint(r.Spec.Endpoint))
	verr.AddErrorAt(path, validateProtocol(r.Spec.Protocol))

	return verr.OrNil()
}

func validateEndpoint(endpoint Endpoint) validators.ValidationError {
	var verr validators.ValidationError

	if endpoint.Address == "" {
		verr.AddViolationAt(validators.RootedAt("address"), validators.MustNotBeEmpty)
	} else if !govalidator.IsIP(endpoint.Address) && !govalidator.IsDNSName(endpoint.Address) {
		verr.AddViolationAt(validators.RootedAt("address"), "address has to be a valid IP or hostname")
	}

	if endpoint.Port == 0 || endpoint.Port > math.MaxUint16 {
		verr.AddViolationAt(validators.RootedAt("port"), "port must be a valid (1-65535)")
	}

	if endpoint.Path != nil && *endpoint.Path != "" {
		pathField := validators.RootedAt("path")
		if !strings.HasPrefix(*endpoint.Path, "/") {
			verr.AddViolationAt(pathField, "must start with /")
		}
		if strings.ContainsAny(*endpoint.Path, "?#") {
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
