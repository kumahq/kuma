package v1alpha1

import (
	"fmt"
	"math"
	"slices"

	"github.com/asaskevich/govalidator"

	datasource_api "github.com/kumahq/kuma/v3/api/common/v1alpha1/datasource"
	common_tls "github.com/kumahq/kuma/v3/api/common/v1alpha1/tls"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

var (
	allMatchProtocols    = core_meta.ProtocolList{core_meta.ProtocolTCP, core_meta.ProtocolGRPC, core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2}
	allVerificationModes = []string{string(TLSVerificationSkipSAN), string(TLSVerificationSkipCA), string(TLSVerificationSkipAll), string(TLSVerificationSecured)}
	allSANMatchTypes     = []string{string(SANMatchPrefix), string(SANMatchExact)}
)

func (r *MeshExternalServiceResource) validate() error {
	var verr validators.ValidationError

	if meta := r.GetMeta(); meta != nil {
		verr.Add(validators.ValidateRFC1035Name(validators.RootedAt("name"), model.GetDisplayName(meta)))
	}

	path := validators.RootedAt("spec")

	verr.AddErrorAt(path.Field("match"), validateMatch(r.Spec.Match))
	// when extension != nil then it's up to the extension to validate endpoints and tls
	if r.Spec.Extension == nil {
		if r.Spec.Endpoints != nil {
			verr.AddErrorAt(path.Field("endpoints"), validateEndpoints(pointer.Deref(r.Spec.Endpoints)))
		}

		if r.Spec.Tls != nil {
			verr.AddErrorAt(path.Field("tls"), validateTls(r.Spec.Tls))
		}
	} else if r.Spec.Tls != nil && r.Spec.Tls.Verification != nil {
		// an extension owns the rest of the tls validation, but never the data source
		// shape: File/EnvVar are read from the control plane process itself, and the
		// remaining types must still carry their required fields, or loadSecureBytes
		// will silently drop the destination after write-time validation accepted it.
		verr.AddErrorAt(path.Field("tls"), validateVerificationDataSources(r.Spec.Tls.Verification))
	}

	if r.Spec.Extension != nil && r.Spec.Extension.Type == "" {
		verr.AddViolationAt(path.Field("extension").Field("type"), validators.MustNotBeEmpty)
	}

	return verr.OrNil()
}

func validateTls(tls *Tls) validators.ValidationError {
	var verr validators.ValidationError

	if tls.Version != nil {
		verr.AddError(validators.RootedAt("version").String(), common_tls.ValidateVersion(tls.Version))
	}

	if tls.Verification != nil {
		path := validators.RootedAt("verification")
		if tls.Verification.ServerName != nil && !govalidator.IsDNSName(*tls.Verification.ServerName) {
			verr.AddViolationAt(path.Field("serverName"), "must be a valid DNS name")
		}
		// checking "" is for backwards compatibility, should be handled by "base policy" in the future
		if !slices.Contains(append(allVerificationModes, ""), string(tls.Verification.Mode)) {
			verr.AddErrorAt(path.Field("mode"), validators.MakeFieldMustBeOneOfErr("mode", allVerificationModes...))
		}
		for i, san := range pointer.Deref(tls.Verification.SubjectAltNames) {
			if !slices.Contains(allSANMatchTypes, string(san.Type)) {
				verr.AddErrorAt(path.Field("subjectAltNames").Index(i).Field("type"), validators.MakeFieldMustBeOneOfErr("type", allSANMatchTypes...))
			}
		}

		if tls.Verification.ClientCert != nil && tls.Verification.ClientKey == nil {
			verr.AddViolation(path.Field("clientKey").String(), validators.MustBeDefined+" when clientCert is defined")
		}
		if tls.Verification.ClientCert == nil && tls.Verification.ClientKey != nil {
			verr.AddViolation(path.Field("clientCert").String(), validators.MustBeDefined+" when clientKey is defined")
		}

		verr.Add(validateSecureDataSource(path.Field("caCert"), tls.Verification.CaCert))
		verr.Add(validateSecureDataSource(path.Field("clientCert"), tls.Verification.ClientCert))
		verr.Add(validateSecureDataSource(path.Field("clientKey"), tls.Verification.ClientKey))
	}

	return verr
}

func validateVerificationDataSources(verification *Verification) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.RootedAt("verification")
	verr.Add(validateSecureDataSource(path.Field("caCert"), verification.CaCert))
	verr.Add(validateSecureDataSource(path.Field("clientCert"), verification.ClientCert))
	verr.Add(validateSecureDataSource(path.Field("clientKey"), verification.ClientKey))
	return verr
}

// validateSecureDataSourceType rejects the data source types that make the control plane read its
// own filesystem or environment, which a MeshExternalService author must never be able to request.
func validateSecureDataSourceType(path validators.PathBuilder, sds *datasource_api.SecureDataSource) validators.ValidationError {
	var verr validators.ValidationError
	if sds == nil {
		return verr
	}
	switch sds.Type {
	case datasource_api.SecureDataSourceFile, datasource_api.SecureDataSourceEnvVar:
		verr.AddViolationAt(path.Field("type"), validators.MustBeOneOf(string(sds.Type), string(datasource_api.SecureDataSourceSecretRef), string(datasource_api.SecureDataSourceInline)))
	}
	return verr
}

func validateSecureDataSource(path validators.PathBuilder, sds *datasource_api.SecureDataSource) validators.ValidationError {
	var verr validators.ValidationError
	if sds == nil {
		return verr
	}
	if typeErr := validateSecureDataSourceType(path, sds); typeErr.HasViolations() {
		return typeErr
	}
	verr.Add(sds.ValidateSecureDataSource(path))
	return verr
}

func validateMatch(match Match) validators.ValidationError {
	var verr validators.ValidationError
	if match.Type != HostnameGeneratorType && match.Type != "" {
		verr.AddViolation(validators.RootedAt("type").String(), fmt.Sprintf("unrecognized type '%s' - only '%s' is supported", match.Type, HostnameGeneratorType))
	}
	if match.Port == 0 || match.Port > math.MaxUint16 {
		verr.AddViolationAt(validators.RootedAt("port"), "port must be a valid (1-65535)")
	}
	if !allMatchProtocols.Contains(match.Protocol) {
		verr.AddErrorAt(validators.RootedAt("protocol"), validators.MakeFieldMustBeOneOfErr("protocol", allMatchProtocols.Strings()...))
	}

	return verr
}

func validateEndpoints(endpoints []Endpoint) validators.ValidationError {
	var verr validators.ValidationError

	for i, endpoint := range endpoints {
		if govalidator.IsIP(endpoint.Address) {
			if endpoint.Port == 0 || endpoint.Port > math.MaxUint16 {
				verr.AddViolationAt(validators.Root().Index(i).Field("port"), "port must be a valid (1-65535)")
			}
		}

		if govalidator.IsDNSName(endpoint.Address) {
			if endpoint.Port == 0 || endpoint.Port > math.MaxUint16 {
				verr.AddViolationAt(validators.Root().Index(i).Field("port"), "port must be a valid (1-65535)")
			}
		}

		if !govalidator.IsIP(endpoint.Address) && !govalidator.IsDNSName(endpoint.Address) {
			verr.AddViolationAt(validators.Root().Index(i).Field("address"), "address has to be a valid IP or hostname")
		}
	}

	return verr
}
