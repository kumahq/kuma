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

type Identity struct {
	Mesh     string
	Name     string
	Services []string
}

type IdentityCertProvider interface {
	Get(ctx context.Context, requestor Identity) (*core_xds.IdentitySecret, error)
}

func NewIdentityProvider(resourceManager core_manager.ResourceManager, caManagers core_ca.Managers) IdentityCertProvider {
	return &identityCertProvider{
		resourceManager: resourceManager,
		caManagers:      caManagers,
	}
}

type identityCertProvider struct {
	resourceManager core_manager.ResourceManager
	caManagers      core_ca.Managers
}

func (s *identityCertProvider) Get(ctx context.Context, requestor Identity) (*core_xds.IdentitySecret, error) {
	meshName := requestor.Mesh

	meshRes := core_mesh.NewMeshResource()
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

	pair, err := caManager.GenerateDataplaneCert(ctx, meshName, backend, requestor.Services)
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate dataplane cert for mesh: %q backend: %q services: %q", meshName, backend.Name, requestor.Services)
	}

	return &core_xds.IdentitySecret{
		PemCerts: [][]byte{pair.CertPEM},
		PemKey:   pair.KeyPEM,
	}, nil
}
