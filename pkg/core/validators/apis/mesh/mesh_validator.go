package mesh

import (
	"context"
	"reflect"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca/provided"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/validators"
)

type MeshValidator struct {
	ProvidedCaManager provided.ProvidedCaManager
}

func (m *MeshValidator) ValidateCreate(ctx context.Context, name string, resource *core_mesh.MeshResource) error {
	switch resource.Spec.GetMtls().GetCa().GetType().(type) {
	case *mesh_proto.CertificateAuthority_Provided_:
		if err := m.validateProvidedCaRoot(ctx, name); err != nil {
			return err
		}
	}
	return nil
}

func (m *MeshValidator) validateProvidedCaRoot(ctx context.Context, mesh string) error {
	certs, err := m.ProvidedCaManager.GetSigningCerts(ctx, mesh)
	if err != nil {
		verr := validators.ValidationError{}
		verr.AddViolation("mtls.ca.provided", "There are no provided CA for a given mesh")
		return verr.OrNil()
	}
	if len(certs) == 0 {
		verr := validators.ValidationError{}
		verr.AddViolation("mtls.ca.provided", "There are no signing certificate in provided CA for a given mesh")
		return verr.OrNil()
	}
	return nil
}

func (m *MeshValidator) ValidateUpdate(ctx context.Context, previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error {
	if err := m.validateCaChange(previousMesh, newMesh); err != nil {
		return err
	}
	return m.validateProvidedCaRoot(ctx, newMesh.Meta.GetName())
}

func (m *MeshValidator) validateCaChange(previousMesh *core_mesh.MeshResource, newMesh *core_mesh.MeshResource) error {
	verr := validators.ValidationError{}
	if previousMesh.Spec.Mtls.Enabled && newMesh.Spec.Mtls.Enabled && reflect.TypeOf(newMesh.Spec.Mtls.Ca.Type) != reflect.TypeOf(previousMesh.Spec.Mtls.Ca.Type) {
		verr.AddViolation("mtls.ca", "Changing CA when mTLS is enabled is forbidden. Disable mTLS first and then change the CA")
	}
	return verr.OrNil()
}
