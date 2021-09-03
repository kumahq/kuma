package secrets

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
)

type Identity struct {
	Mesh     string
	Name     string
	Services mesh_proto.MultiValueTagSet
}

type IdentityProvider interface {
	// Get returns PEM encoded cert + key, backend that was used to generate this pair and an error.
	Get(context.Context, Identity, *core_mesh.MeshResource) (*core_xds.IdentitySecret, string, error)
}

func NewIdentityProvider(caManagers core_ca.Managers) IdentityProvider {
	return &identityCertProvider{
		caManagers: caManagers,
	}
}

type identityCertProvider struct {
	caManagers core_ca.Managers
}

func (s *identityCertProvider) Get(ctx context.Context, requestor Identity, mesh *core_mesh.MeshResource) (*core_xds.IdentitySecret, string, error) {
	backend := mesh.GetEnabledCertificateAuthorityBackend()
	if backend == nil {
		return nil, "", errors.Errorf("CA default backend in mesh %q has to be defined", mesh.GetMeta().GetName())
	}

	caManager, exist := s.caManagers[backend.Type]
	if !exist {
		return nil, "", errors.Errorf("CA manager of type %s not exist", backend.Type)
	}

	pair, err := caManager.GenerateDataplaneCert(ctx, mesh.GetMeta().GetName(), backend, requestor.Services)
	if err != nil {
		return nil, "", errors.Wrapf(err, "could not generate dataplane cert for mesh: %q backend: %q services: %q", mesh.GetMeta().GetName(), backend.Name, requestor.Services)
	}

	return &core_xds.IdentitySecret{
		PemCerts: [][]byte{pair.CertPEM},
		PemKey:   pair.KeyPEM,
	}, backend.Name, nil
}
