// +kubebuilder:object:generate=true
package tls

import (
	"slices"

	"github.com/kumahq/kuma/pkg/core/validators"
)

// +kubebuilder:validation:Enum=TLSAuto;TLS10;TLS11;TLS12;TLS13
type TlsVersion string

const (
	TLSVersionAuto TlsVersion = "TLSAuto"
	TLSVersion10   TlsVersion = "TLS10"
	TLSVersion11   TlsVersion = "TLS11"
	TLSVersion12   TlsVersion = "TLS12"
	TLSVersion13   TlsVersion = "TLS13"
)

var TlsVersionOrder = map[TlsVersion]int{
	TLSVersion10: 0,
	TLSVersion11: 1,
	TLSVersion12: 2,
	TLSVersion13: 3,
}

var allTlsVersions = []string{string(TLSVersionAuto), string(TLSVersion10), string(TLSVersion11), string(TLSVersion12), string(TLSVersion13)}

type Version struct {
	// Min defines minimum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`.
	// +kubebuilder:default=TLSAuto
	Min *TlsVersion `json:"min,omitempty"`
	// Max defines maximum supported version. One of `TLSAuto`, `TLS10`, `TLS11`, `TLS12`, `TLS13`.
	// +kubebuilder:default=TLSAuto
	Max *TlsVersion `json:"max,omitempty"`
}

func ValidateVersion(version *Version) validators.ValidationError {
	var verr validators.ValidationError
	path := validators.Root()
	specificMin := false
	specificMax := false
	if version.Min != nil {
		if !slices.Contains(allTlsVersions, string(*version.Min)) {
			verr.AddErrorAt(path.Field("min"), validators.MakeFieldMustBeOneOfErr("min", allTlsVersions...))
		} else if *version.Min != TLSVersionAuto {
			specificMin = true
		}
	}
	if version.Max != nil {
		if !slices.Contains(allTlsVersions, string(*version.Max)) {
			verr.AddErrorAt(path.Field("max"), validators.MakeFieldMustBeOneOfErr("max", allTlsVersions...))
		} else if *version.Max != TLSVersionAuto {
			specificMax = true
		}
	}

	if specificMin && specificMax && TlsVersionOrder[*version.Min] > TlsVersionOrder[*version.Max] {
		verr.AddViolationAt(path.Field("min"), "min version must be lower than max")
	}

	return verr
}
