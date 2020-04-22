package identity

import (
	"context"

	"github.com/pkg/errors"

	core_ca "github.com/Kong/kuma/pkg/core/ca"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	sds_provider "github.com/Kong/kuma/pkg/sds/provider"
)

func New(resourceManager core_manager.ResourceManager, caManagers core_ca.CaManagers) sds_provider.SecretProvider {
	return &identityCertProvider{
		resourceManager: resourceManager,
		caManagers:      caManagers,
	}
}

type identityCertProvider struct {
	resourceManager core_manager.ResourceManager
	caManagers      core_ca.CaManagers
}

func (s *identityCertProvider) RequiresIdentity() bool {
	return true
}

func (s *identityCertProvider) Get(ctx context.Context, name string, requestor sds_auth.Identity) (sds_provider.Secret, error) {
	meshName := requestor.Mesh

	meshRes := &core_mesh.MeshResource{}
	if err := s.resourceManager.Get(ctx, meshRes, core_store.GetByKey(meshName, meshName)); err != nil {
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

	pair, err := caManager.GenerateDataplaneCert(ctx, meshName, *backend, requestor.Service)
	if err != nil {
		return nil, errors.Wrapf(err, "could not generate dataplane cert for mesh: %q backend: %q service: %q", meshName, backend.Name, requestor.Service)
	}

	return &IdentityCertSecret{
		PemCerts: [][]byte{pair.CertPEM},
		PemKey:   pair.KeyPEM,
	}, nil
}
