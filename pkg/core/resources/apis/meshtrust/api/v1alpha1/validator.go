package v1alpha1

import (
	"encoding/pem"
	"fmt"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *MeshTrustResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	if r.Spec.Origin != nil {
		verr.AddViolationAt(path.Field("origin"), "cannot be provided by the user")
	}
	verr.Add(validateCABundles(path.Field("caBundles"), r.Spec.CABundles))
	verr.Add(validateTrustDomain(path, r.Spec.TrustDomain))
	return verr.OrNil()
}

func validateTrustDomain(path validators.PathBuilder, td string) validators.ValidationError {
	var verr validators.ValidationError
	if td == "" {
		verr.AddViolationAt(path.Field("trustDomain"), "trustDomain needs to be defined")
	}
	if len(td) > 253 {
		verr.AddViolationAt(path.Field("trustDomain"), fmt.Sprintf("trustDomain needs to shorter than 253 sings, current has %d", len(td)))
	}
	return verr
}

func validateCABundles(path validators.PathBuilder, bundles []CABundle) validators.ValidationError {
	var verr validators.ValidationError
	if len(bundles) == 0 {
		verr.AddViolationAt(path, "atleast 1 bundle needs to be provided")
		return verr
	}
	for i, bundle := range bundles {
		switch bundle.Type {
		case PemCABundleType:
			path := path.Index(i).Field("pem")
			if bundle.Pem == nil {
				verr.AddViolationAt(path, "pem needs to be defined")
				continue
			}
			if !isPEMFormat(bundle.Pem.Value) {
				verr.AddViolationAt(path.Field("value"), "provided certificate has incorrect format")
			}
		}
	}
	return verr
}

func isPEMFormat(s string) bool {
	block, _ := pem.Decode([]byte(s))
	return block != nil
}
