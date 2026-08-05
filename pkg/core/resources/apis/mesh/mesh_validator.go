package mesh

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

var AllowedMTLSBackends = 1

func (m *MeshResource) Validate() error {
	var verr validators.ValidationError
	verr.AddError("mtls", validateMtls(m.Spec.Mtls))
	return verr.OrNil()
}

func validateMtls(mtls *mesh_proto.Mesh_Mtls) validators.ValidationError {
	var verr validators.ValidationError
	if mtls == nil {
		return verr
	}
	if len(mtls.GetBackends()) > AllowedMTLSBackends {
		verr.AddViolationAt(validators.RootedAt("backends"), fmt.Sprintf("cannot have more than %d backends", AllowedMTLSBackends))
	}

	usedNames := map[string]bool{}
	for i, backend := range mtls.GetBackends() {
		if usedNames[backend.Name] {
			verr.AddViolationAt(validators.RootedAt("backends").Index(i).Field("name"), fmt.Sprintf("%q name is already used for another backend", backend.Name))
		}
		usedNames[backend.Name] = true

		if backend.GetDpCert() != nil {
			_, err := ParseDuration(backend.GetDpCert().GetRotation().GetExpiration())
			if err != nil {
				verr.AddViolation("dpcert.rotation.expiration", "has to be a valid format")
			}
		}
	}
	if mtls.GetEnabledBackend() != "" && !usedNames[mtls.GetEnabledBackend()] {
		verr.AddViolation("enabledBackend", "has to be set to one of the backends in the mesh")
	}
	return verr
}
