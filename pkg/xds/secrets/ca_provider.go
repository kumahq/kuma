package secrets

import (
	"context"

	"github.com/pkg/errors"

	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type CaProvider interface {
	// Get returns all PEM encoded CAs, a list of CAs that were used to generate a secret and an error.
	Get(context.Context, *core_mesh.MeshResource) (*core_xds.CaSecret, []string, error)
}

func NewCaProvider(caManagers core_ca.Managers) CaProvider {
	return &meshCaProvider{
		caManagers: caManagers,
	}
}

type meshCaProvider struct {
	caManagers core_ca.Managers
}

func (s *meshCaProvider) Get(ctx context.Context, mesh *core_mesh.MeshResource) (*core_xds.CaSecret, []string, error) {
	backend := mesh.GetEnabledCertificateAuthorityBackend()
	if backend == nil {
		return nil, nil, errors.New("CA backend is nil")
	}

	caManager, exist := s.caManagers[backend.Type]
	if !exist {
		return nil, nil, errors.Errorf("CA manager of type %s not exist", backend.Type)
	}

	certs, err := caManager.GetRootCert(ctx, mesh.GetMeta().GetName(), backend)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not get root certs")
	}

	return &core_xds.CaSecret{
		PemCerts: certs,
	}, []string{backend.Name}, nil
}
