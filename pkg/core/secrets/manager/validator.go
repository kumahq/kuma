package manager

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
)

var log = core.Log.WithName("secrets").WithName("validator")

type SecretValidator interface {
	ValidateDelete(ctx context.Context, secretName string, secretMesh string) error
}

type ValidateDelete func(ctx context.Context, secretName string, secretMesh string) error

func (f ValidateDelete) ValidateDelete(ctx context.Context, secretName string, secretMesh string) error {
	return f(ctx, secretName, secretMesh)
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
	meshRes := core_mesh.NewMeshResource()
	err := s.store.Get(ctx, meshRes, core_store.GetByKey(mesh, model.NoMesh))
	if err != nil {
		if core_store.IsNotFound(err) {
			return nil // when Mesh no longer exist we should be able to safely delete a secret because it's not referenced anywhere
		}
		return err
	}

	var verr validators.ValidationError
	for _, backend := range meshRes.Spec.GetMtls().GetBackends() {
		used, err := s.secretUsedByMTLSBackend(name, meshRes.GetMeta().GetName(), backend)
		if err != nil {
			log.Info("error while checking if secret is used by mTLS backend. Deleting secret anyway", "cause", err)
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
	secrets, err := caManager.UsedSecrets(mesh, backend)
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
