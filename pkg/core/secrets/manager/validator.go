package manager

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	"github.com/Kong/kuma/pkg/core/resources/model"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/validators"
)

// MeshAccessor is used instead of MeshManager, otherwise we would have cyclic dependency.
// MeshManager depends on SecretManager which depends on SecretValidator which would depend on ResourceManager
type MeshAccessor = func(context.Context, model.Resource, ...core_store.GetOptionsFunc) error

type SecretValidator interface {
	ValidateDelete(ctx context.Context, secretName string, secretMesh string) error
}

func NewSecretValidator(caManagers ca.Managers, meshAccessor MeshAccessor) SecretValidator {
	return &secretValidator{
		caManagers:   caManagers,
		meshAccessor: meshAccessor,
	}
}

type secretValidator struct {
	caManagers   ca.Managers
	meshAccessor MeshAccessor
}

func (s *secretValidator) ValidateDelete(ctx context.Context, secretName string, secretMesh string) error {
	mesh := &mesh_core.MeshResource{}
	err := s.meshAccessor(ctx, mesh, core_store.GetByKey(secretMesh, secretMesh))
	if err != nil {
		return err
	}

	var verr validators.ValidationError
	for _, backend := range mesh.Spec.GetMtls().GetBackends() {
		used, err := s.secretUsedByMTLSBackend(secretName, mesh.GetMeta().GetName(), backend)
		if err != nil {
			return err
		}
		if used {
			verr.AddViolation("name", fmt.Sprintf(`The secret %q that you are trying to remove is currently in use in Mesh %q in mTLS backend %q. Please remove the reference from the %q backend before removing the secret.`, secretName, secretMesh, backend.Name, backend.Name))
		}
	}
	return verr.OrNil()
}

func (s *secretValidator) secretUsedByMTLSBackend(secret string, mesh string, backend *mesh_proto.CertificateAuthorityBackend) (bool, error) {
	caManager := s.caManagers[backend.Type]
	if caManager == nil { // this should be caught earlier by validator
		return false, errors.Errorf("manager of type %q does not exist", backend.Type)
	}
	secrets, err := caManager.UsedSecrets(mesh, *backend)
	if err != nil {
		return false, errors.Wrapf(err, "could not retrieve secrets in use by backend %q", backend.Name)
	}
	for _, secretName := range secrets {
		if secretName == secret {
			return true, nil
		}
	}
	return false, nil
}
