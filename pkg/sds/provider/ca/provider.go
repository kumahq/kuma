package ca

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
	return &meshCaProvider{
		resourceManager: resourceManager,
		caManagers:      caManagers,
	}
}

type meshCaProvider struct {
	resourceManager core_manager.ResourceManager
	caManagers      core_ca.Managers
}

func (s *meshCaProvider) RequiresIdentity() bool {
	return false
}

func (s *meshCaProvider) Get(ctx context.Context, resource string, requestor sds_provider.Identity) (sds_provider.Secret, error) {
	meshName := requestor.Mesh

	meshRes := core_mesh.NewMeshResource()
	if err := s.resourceManager.Get(ctx, meshRes, core_store.GetByKey(meshName, model.NoMesh)); err != nil {
		return nil, errors.Wrapf(err, "failed to find a Mesh %q", meshName)
	}

	backend := meshRes.GetEnabledCertificateAuthorityBackend()
	if backend == nil {
		return nil, errors.New("CA backend is nil")
	}

	caManager, exist := s.caManagers[backend.Type]
	if !exist {
		return nil, errors.Errorf("CA manager of type %s not exist", backend.Type)
	}

	certs, err := caManager.GetRootCert(ctx, meshName, backend)
	if err != nil {
		return nil, errors.Wrap(err, "could not get root certs")
	}

	return &MeshCaSecret{
		PemCerts: certs,
	}, nil
}
