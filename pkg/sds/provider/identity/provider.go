package identity

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	builtin_ca "github.com/Kong/kuma/pkg/core/ca/builtin"
	provided_ca "github.com/Kong/kuma/pkg/core/ca/provided"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/tls"

	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	sds_provider "github.com/Kong/kuma/pkg/sds/provider"
)

func New(resourceManager core_manager.ResourceManager, builtinCaManager builtin_ca.BuiltinCaManager, providedCaManager provided_ca.ProvidedCaManager) sds_provider.SecretProvider {
	return &identityCertProvider{
		resourceManager:   resourceManager,
		builtinCaManager:  builtinCaManager,
		providedCaManager: providedCaManager,
	}
}

type identityCertProvider struct {
	resourceManager   core_manager.ResourceManager
	builtinCaManager  builtin_ca.BuiltinCaManager
	providedCaManager provided_ca.ProvidedCaManager
}

func (s *identityCertProvider) RequiresIdentity() bool {
	return true
}

func (s *identityCertProvider) Get(ctx context.Context, name string, requestor sds_auth.Identity) (sds_provider.Secret, error) {
	meshName := requestor.Mesh
	list := &core_mesh.MeshResourceList{}
	if err := s.resourceManager.List(ctx, list, core_store.ListByMesh(meshName)); err != nil {
		return nil, errors.Wrapf(err, "failed to find a Mesh %q", meshName)
	}
	if len(list.Items) == 0 {
		return nil, errors.Errorf("there is no Mesh %q", meshName)
	}
	if len(list.Items) != 1 {
		return nil, errors.Errorf("there are multiple Meshes named %q", meshName)
	}
	mesh := list.Items[0]

	var generator func(context.Context, string, string) (*tls.KeyPair, error)
	switch mesh.Spec.GetMtls().GetCa().GetType().(type) {
	case *mesh_proto.CertificateAuthority_Builtin_:
		generator = s.builtinCaManager.GenerateWorkloadCert
	case *mesh_proto.CertificateAuthority_Provided_:
		generator = s.providedCaManager.GenerateWorkloadCert
	default:
		return nil, errors.Errorf("Mesh %q has unsupported CA type", meshName)
	}

	workloadCert, err := generator(ctx, mesh.Meta.GetName(), requestor.Service)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to generate a Workload Identity Certificate for %+v", requestor)
	}
	return &IdentityCertSecret{
		PemCerts: [][]byte{workloadCert.CertPEM},
		PemKey:   workloadCert.KeyPEM,
	}, nil
}
