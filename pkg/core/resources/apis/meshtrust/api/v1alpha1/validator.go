package v1alpha1

import (
	"encoding/pem"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshTrustResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.Add(validateCABundles(path.Field("caBundles"), r.Spec.CABundles))
	verr.Add(validateTrustDomain(path, r.Spec.TrustDomain))
	return verr.OrNil()
}

func validateTrustDomain(path validators.PathBuilder, td string) validators.ValidationError {
	var verr validators.ValidationError
	if td == "" {
		verr.AddViolationAt(path.Field("trustDomain"), "trustDomain needs to be defined")
	}
	verr.Add(validators.ValidateLength(path.Field("trustDomain"), 253, td))
	return verr
}

func validateCABundles(path validators.PathBuilder, bundles []CABundle) validators.ValidationError {
	var verr validators.ValidationError
	if len(bundles) == 0 {
		verr.AddViolationAt(path, validators.MustNotBeEmpty)
		return verr
	}
	for i, bundle := range bundles {
		switch bundle.Type {
		case PemCABundleType:
			path := path.Index(i).Field("pem")
			if bundle.PEM == nil {
				verr.AddViolationAt(path, validators.MustBeDefined)
				continue
			}
			if !isPEMCertificate(bundle.PEM.Value) {
				verr.AddViolationAt(path.Field("value"), "provided certificate has incorrect format")
			}
		default:
			verr.AddViolationAt(path.Field("type"), validators.MustBeOneOf(string(bundle.Type), string(PemCABundleType)))
		}
	}
	return verr
}

func isPEMCertificate(s string) bool {
	block, _ := pem.Decode([]byte(s))
	if block == nil {
		return false
	}

	switch block.Type {
	case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY", "DSA PRIVATE KEY", "ENCRYPTED PRIVATE KEY":
		return false
	default:
		return true
	}
}
