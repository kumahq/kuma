package manager

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core"
	"github.com/Kong/kuma/pkg/core/ca"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/core/validators"
)

var log = core.Log.WithName("secrets").WithName("validator")

type SecretValidator interface {
	ValidateDelete(ctx context.Context, secretName string, secretMesh string) error
}

func NewSecretValidator(caManagers ca.Managers, store core_store.ResourceStore) SecretValidator {
	return &secretValidator{
		caManagers: caManagers,
		store:      store,
	}
}

type secretValidator struct {
	caManagers ca.Managers
	store      core_store.ResourceStore
}

func (s *secretValidator) ValidateDelete(ctx context.Context, name string, mesh string) error {
	meshRes := &mesh_core.MeshResource{}
	err := s.store.Get(ctx, meshRes, core_store.GetByKey(mesh, mesh))
	if err != nil {
		return err
	}

	var verr validators.ValidationError
	for _, backend := range meshRes.Spec.GetMtls().GetBackends() {
		used, err := s.secretUsedByMTLSBackend(name, meshRes.GetMeta().GetName(), backend)
		if err != nil {
			log.Info("Error while checking if secret is used by mTLS backend. Deleting secret anyway", err)
		}
		if used {
			verr.AddViolation("name", fmt.Sprintf(`The secret %q that you are trying to remove is currently in use in Mesh %q in mTLS backend %q. Please remove the reference from the %q backend before removing the secret.`, name, mesh, backend.Name, backend.Name))
		}
	}
	return verr.OrNil()
}

func (s *secretValidator) secretUsedByMTLSBackend(name string, mesh string, backend *mesh_proto.CertificateAuthorityBackend) (bool, error) {
	caManager := s.caManagers[backend.Type]
	if caManager == nil { // this should be caught earlier by validator
		return false, errors.Errorf("manager of type %q does not exist", backend.Type)
	}
	secrets, err := caManager.UsedSecrets(mesh, *backend)
	if err != nil {
		return false, errors.Wrapf(err, "could not retrieve secrets in use by backend %q", backend.Name)
	}
	for _, secret := range secrets {
		if secret == name {
			return true, nil
		}
	}
	return false, nil
}
