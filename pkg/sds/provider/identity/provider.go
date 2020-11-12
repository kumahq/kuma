package identity

import (
	"context"

	"github.com/pkg/errors"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	sds_provider "github.com/kumahq/kuma/pkg/sds/provider"
)

func New(resourceManager core_manager.ResourceManager, caManagers core_ca.Managers) sds_provider.SecretProvider {
	return &identityCertProvider{
		resourceManager: resourceManager,
		caManagers:      caManagers,
	}
}

type identityCertProvider struct {
	resourceManager core_manager.ResourceManager
	caManagers      core_ca.Managers
}

func (s *identityCertProvider) RequiresIdentity() bool {
	return true
}

func (s *identityCertProvider) Get(ctx context.Context, name string, requestor sds_provider.Identity) (sds_provider.Secret, error) {
	meshName := requestor.Mesh

	meshRes := &core_mesh.MeshResource{}
	if err := s.resourceManager.Get(ctx, meshRes, core_store.GetByKey(meshName, model.NoMesh)); err != nil {
		return nil, errors.Wrapf(err, "failed to find a Mesh %q", meshName)
	}

	backend := meshRes.GetEnabledCertificateAuthorityBackend()
	if backend == nil {
		return nil, errors.Errorf("CA default backend in mesh %q has to be defined", meshName)
	}

	caManager, exist := s.caManagers[backend.Type]
	if !exist {
		return nil, errors.Errorf("CA manager of type %s not exist", backend.Type)
	}

	pair, err := caManager.GenerateDataplaneCert(ctx, meshName, *backend, requestor.Services)
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate dataplane cert for mesh: %q backend: %q services: %q", meshName, backend.Name, requestor.Services)
	}

	return &IdentityCertSecret{
		PemCerts: [][]byte{pair.CertPEM},
		PemKey:   pair.KeyPEM,
	}, nil
}
