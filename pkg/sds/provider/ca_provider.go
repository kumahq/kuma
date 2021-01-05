package provider

import (
	"context"

	"github.com/pkg/errors"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type CaProvider interface {
	Get(ctx context.Context, mesh string) (*core_xds.CaSecret, error)
}

func NewCaProvider(resourceManager core_manager.ResourceManager, caManagers core_ca.Managers) CaProvider {
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

func (s *meshCaProvider) Get(ctx context.Context, mesh string) (*core_xds.CaSecret, error) {
	meshRes := core_mesh.NewMeshResource()
	if err := s.resourceManager.Get(ctx, meshRes, core_store.GetByKey(mesh, model.NoMesh)); err != nil {
		return nil, errors.Wrapf(err, "failed to find a Mesh %q", mesh)
	}

	backend := meshRes.GetEnabledCertificateAuthorityBackend()
	if backend == nil {
		return nil, errors.New("CA backend is nil")
	}

	caManager, exist := s.caManagers[backend.Type]
	if !exist {
		return nil, errors.Errorf("CA manager of type %s not exist", backend.Type)
	}

	certs, err := caManager.GetRootCert(ctx, mesh, backend)
	if err != nil {
		return nil, errors.Wrap(err, "could not get root certs")
	}

	return &core_xds.CaSecret{
		PemCerts: certs,
	}, nil
}
