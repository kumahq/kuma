package v1alpha1

import (
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/kumahq/kuma/pkg/core/validators"
	"math"
	"slices"
	"strings"
)

var allMatchProtocols = []string{string(TcpProtocol), string(GrpcProtocol), string(HttpProtocol), string(Http2Protocol)}
var allTlsVersions = []string{string(TLSVersionAuto), string(TLSVersion10), string(TLSVersion11), string(TLSVersion12), string(TLSVersion13)}
var allVerificationModes = []string{string(TLSVerificationSkipSAN), string(TLSVerificationSkipCA), string(TLSVerificationSkipAll), string(TLSVerificationSecured)}
var allSANMatchTypes = []string{string(SANMatchPrefix), string(SANMatchExact)}

func (r *MeshExternalServiceResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")

	verr.AddErrorAt(path.Field("match"), validateMatch(r.Spec.Match))
	// when extension != nil then it's up to the extension to validate endpoints and tls
	if r.Spec.Extension == nil {
		if r.Spec.Endpoints != nil {
			verr.AddErrorAt(path.Field("endpoints"), validateEndpoints(r.Spec.Endpoints))
		}

		if r.Spec.Tls != nil {
			verr.AddErrorAt(path.Field("tls"), validateTls(r.Spec.Tls))
		}
	} else {
		if r.Spec.Extension.Type == "" {
			verr.AddViolationAt(path.Field("extension").Field("type"), validators.MustNotBeEmpty)
		}
	}

	return verr.OrNil()
}

func validateTls(tls *Tls) validators.ValidationError {
	var verr validators.ValidationError

	if tls.Version != nil {
		path := validators.RootedAt("version")
		if tls.Version.Min != nil {
			if !slices.Contains(allTlsVersions, string(*tls.Version.Min)) {
				verr.AddErrorAt(path.Field("min"), validators.MakeFieldMustBeOneOfErr("min", allTlsVersions...))
			}
		}
		if tls.Version.Max != nil {
			if !slices.Contains(allTlsVersions, string(*tls.Version.Max)) {
				verr.AddErrorAt(path.Field("max"), validators.MakeFieldMustBeOneOfErr("max", allTlsVersions...))
			}
		}
	}

	if tls.Verification != nil {
		path := validators.RootedAt("verification")

		if tls.Verification.Mode != nil {
			if !slices.Contains(allVerificationModes, string(*tls.Verification.Mode)) {
				verr.AddErrorAt(path.Field("mode"), validators.MakeFieldMustBeOneOfErr("mode", allVerificationModes...))
			}
		}

		if tls.Verification.SubjectAltNames != nil {
			for i, san := range *tls.Verification.SubjectAltNames {
				if !slices.Contains(allSANMatchTypes, string(san.Type)) {
					verr.AddErrorAt(path.Field("subjectAltNames").Index(i).Field("type"), validators.MakeFieldMustBeOneOfErr("type", allSANMatchTypes...))
				}
			}
		}

		if tls.Verification.ClientCert != nil && tls.Verification.ClientKey == nil {
			verr.AddViolation(path.Field("clientKey").String(), validators.MustBeDefined + "when clientCert is defined")
		}
		if tls.Verification.ClientCert == nil && tls.Verification.ClientKey != nil {
			verr.AddViolation(path.Field("clientCert").String(), validators.MustBeDefined + "when clientKey is defined")
		}
	}

	return verr
}

func validateMatch(match Match) validators.ValidationError {
	var verr validators.ValidationError
	if match.Type != HostnameGeneratorType {
		verr.AddViolation(validators.RootedAt("type").String(), fmt.Sprintf("unrecognized type '%s' - only '%s' is supported", match.Type, HostnameGeneratorType))
	}
	if match.Port == 0 || match.Port > math.MaxUint16 {
		verr.AddViolationAt(validators.RootedAt("port"), "port must be a valid (1-65535)")
	}
	if !slices.Contains(allMatchProtocols, string(match.Protocol)) {
		verr.AddErrorAt(validators.RootedAt("protocol"), validators.MakeFieldMustBeOneOfErr("protocol", allMatchProtocols...))
	}

	return verr
}

func validateEndpoints(endpoints []Endpoint) validators.ValidationError {
	var verr validators.ValidationError

	for i, endpoint := range endpoints {
		if endpoint.Port == 0 || endpoint.Port > math.MaxUint16 {
			verr.AddViolationAt(validators.Root().Index(i).Field("port"), "port must be a valid (1-65535)")
		}

		if !(govalidator.IsIP(endpoint.Address) || govalidator.IsDNSName(endpoint.Address) || isValidUnixPath(endpoint.Address)) {
			verr.AddViolationAt(validators.Root().Index(i).Field("address"), "address has to be a valid IP or hostname or a unix path")
		}
	}

	return verr
}

func isValidUnixPath(path string) bool {
	if strings.HasPrefix(path, "unix://") {
		parts := strings.Split(path, "unix://")
		filePath := parts[1]
		return govalidator.IsUnixFilePath(filePath)
	} else {
		return false
	}
}
